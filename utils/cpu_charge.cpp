#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <string>

#include <fcntl.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/un.h>
#include <unistd.h>

/* lit une ligne sur fd, puis la stocke dans res
renvoie 0 si la lecture est terminée, et 1 sinon */
int readline(int fd, std::string &res) {
    res.clear();
    char c;

    while(1) {
        ssize_t r = read(fd, &c, 1);
        if(r < 0) {
            std::cout << "readline:erreur de lecture (readline)" << std::endl;
            return 0;
        }

        if(r == 0) {
            if (res.empty()) return 0;
            else return 1;
        }

        if(c == '\n')
            return 1;

        res += c;
    }
}

/* renvoie l'indice du premier entier dans une chaine de caractère */
int get_indice_of_int(const char *str) {
    for(size_t i = 0; i < strlen(str); i++) {
        if(str[i] >= '0' && str[i] <= '9')
            return i;
    }
    return -1;
}

/* renvoie l'indice du premier espace dans une chaine de caractère */
int get_indice_of_space(const char *str) {
    for(size_t i = 0; i < strlen(str); i++) {
        if(str[i] == ' ')
            return i;
    }
    return -1;
}

/* renvoie le premier entier dans une chaine de caractère sous forme de long */
long get_long_in_str(const char *str) {
    int ind = get_indice_of_int(str);
    return strtol(&str[ind], NULL, 10);
}

/* renvoie le premier nombre réel dans une chaine de caractère sous forme de double */
double get_double_in_str(const char *str) {
    int ind = get_indice_of_int(str);
    return strtod(&str[ind], NULL);
}

/* on traite de le fichier /proc/meminfo pour récupérer la valeur de MemAvailable (= mémoire disponible) */
void get_mem_info(int fd_meminfo, unsigned long *mem) {
    std::string str;
    while(readline(fd_meminfo, str)) {
        const char *c = str.c_str();
        if(strncmp(c, "MemAvailable:", strlen("MemAvailable:")) != 0)
            continue;
        else {
            *mem = get_long_in_str(c);
            lseek(fd_meminfo, 0, SEEK_SET);
            break;
        }
    }
}

/* on récupère les infos des cpu (nombre de coeurs, somme de la fréquence des coeurs) */
void get_cpu_info(int fd_cpuinfo, double &freqs_sum, unsigned int &nb_coeur) {
    std::string str;
    while(readline(fd_cpuinfo, str)) {
        const char *c = str.c_str();
        if(strncmp(c, "cpu MHz", strlen("cpu MHz")) != 0)
            continue;
        else {
            freqs_sum += get_double_in_str(c);
            nb_coeur++;
        }
    }
}

/* parse la chaine de caractère pour en récupérer les infos (infos de /proc/stat) */
void get_cpu_stat(unsigned long cpu_stat[10], std::string &str) {
    const char *c_str = str.c_str();
    int space = 0;
    for(int i = 0; i < 10; i++) {
        int ind = space + get_indice_of_int(c_str + space);
        cpu_stat[i] = strtol(&str[ind], NULL, 10);
        space = ind + get_indice_of_space(c_str + ind);
    }
}

/* on récupère les stats du cpu 2 fois, avec une attente de 2sec entre les deux lectures */
void parse_cpu_stats(int fd_stat, unsigned long cpu_stat[10], unsigned long cpu_stat2[10]) {
    std::string str;

    readline(fd_stat, str);
    get_cpu_stat(cpu_stat, str);

    sleep(2);
    lseek(fd_stat, 0, SEEK_SET);

    readline(fd_stat, str);
    get_cpu_stat(cpu_stat2, str);
}

double get_cpu_idle_percentage(unsigned long cpu_stat[10], unsigned long cpu_stat2[10]) {
    unsigned long idle_iowait1 = cpu_stat[3] + cpu_stat[4];
    unsigned long idle_iowait2 = cpu_stat2[3] + cpu_stat2[4];
    double delta_idle_iowait = idle_iowait2 - idle_iowait1;

    unsigned long total1 = 0;
    for(int i = 0; i < 10; i++)
        total1 += cpu_stat[i];

    unsigned long total2 = 0;
    for(int i = 0; i < 10; i++)
        total2 += cpu_stat2[i];

    double delta_total = total2 - total1;

    return (delta_idle_iowait / delta_total) * 100;
}

int main(int argc, char *argv[]) {
    int fd_meminfo = open("/proc/meminfo", O_RDONLY);
    int fd_cpuinfo = open("/proc/cpuinfo", O_RDONLY);
    int fd_stat = open("/proc/stat", O_RDONLY);

    unsigned long mem_available = 0; // mémoire disponible en kB
    unsigned int nb_coeur = 0; // nb de coeur du cpu
    double freqs_sum = 0; // somme des fréquences des coeurs du cpu

    // on récupère la valeur de memAvailable
    get_mem_info(fd_meminfo, &mem_available);

    // on récupère le nombre de coeur et la somme des fréquences
    get_cpu_info(fd_cpuinfo, freqs_sum, nb_coeur);

    // on récupère les stats du cpu 2 fois
    unsigned long cpu_stat[10], cpu_stat2[10];
    parse_cpu_stats(fd_stat, cpu_stat, cpu_stat2);

    double idle_percentage = get_cpu_idle_percentage(cpu_stat, cpu_stat2);
    double freqs_by_idle = freqs_sum * (idle_percentage/100);

    std::cout << "idle % : " << idle_percentage << std::endl;
    std::cout << "mem_available : " << mem_available << std::endl;
    std::cout << "freqs * cpu_idle = " << freqs_by_idle << std::endl;

    int sock = socket(AF_UNIX, SOCK_STREAM, 0);

    sockaddr_un addr;
    addr.sun_family = AF_UNIX;
    strcpy(addr.sun_path, "/tmp/cpu.sock");

    connect(sock, (sockaddr*)&addr, sizeof(addr));

    while(1) {
        get_mem_info(fd_meminfo, &mem_available);
        get_cpu_info(fd_cpuinfo, freqs_sum, nb_coeur);
        parse_cpu_stats(fd_stat, cpu_stat, cpu_stat2); // Attention, fait un sleep(2) qui crée un délai
        double idle_percentage = get_cpu_idle_percentage(cpu_stat, cpu_stat2);
        double freqs_by_idle = freqs_sum * (idle_percentage/100);

        write(sock, &mem_available, sizeof(mem_available));
        write(sock, &freqs_by_idle, sizeof(freqs_by_idle));
    }

    close(sock);
    close(fd_cpuinfo);
    close(fd_meminfo);
    close(fd_stat);

    return 0;
}
