#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <cstdint>
#include <iostream>
#include <string>

#include <errno.h>
#include <fcntl.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/un.h>
#include <unistd.h>

/* lit une ligne sur fd, puis la stocke dans res
   renvoie 0 si la lecture est terminée (EOF + ligne vide), 1 sinon */
int readline(int fd, std::string &res) {
    res.clear();
    char c;

    while (1) {
        ssize_t r = read(fd, &c, 1);
        if (r < 0) {
            if (errno == EINTR) continue;
            std::cerr << "readline: erreur de lecture" << std::endl;
            return 0;
        }
        if (r == 0) {
            // EOF : on renvoie 1 si on a déjà lu quelque chose, 0 sinon
            return res.empty() ? 0 : 1;
        }
        if (c == '\n') return 1;
        res += c;
    }
}

/* indice du premier chiffre dans la chaîne, -1 sinon */
int get_indice_of_int(const char *str) {
    for (size_t i = 0; i < strlen(str); i++) {
        if (str[i] >= '0' && str[i] <= '9') return (int)i;
    }
    return -1;
}

/* indice du premier espace dans la chaîne, -1 sinon */
int get_indice_of_space(const char *str) {
    for (size_t i = 0; i < strlen(str); i++) {
        if (str[i] == ' ') return (int)i;
    }
    return -1;
}

long get_long_in_str(const char *str) {
    int ind = get_indice_of_int(str);
    if (ind < 0) return 0;
    return strtol(&str[ind], NULL, 10);
}

double get_double_in_str(const char *str) {
    int ind = get_indice_of_int(str);
    if (ind < 0) return 0.0;
    return strtod(&str[ind], NULL);
}

/* lit /proc/meminfo et récupère MemAvailable (en kB)
   on lseek systématiquement pour pouvoir relire à chaque itération */
void get_mem_info(int fd_meminfo, uint64_t *mem) {
    lseek(fd_meminfo, 0, SEEK_SET);
    std::string str;
    while (readline(fd_meminfo, str)) {
        const char *c = str.c_str();
        if (strncmp(c, "MemAvailable:", strlen("MemAvailable:")) == 0) {
            *mem = (uint64_t)get_long_in_str(c);
            return;
        }
    }
}

/* lit /proc/cpuinfo : nombre de coeurs et somme des fréquences (MHz)
   IMPORTANT : on remet freqs_sum et nb_coeur à 0 pour ne pas accumuler
   entre les appels, et on lseek le fd au début */
void get_cpu_info(int fd_cpuinfo, double &freqs_sum, unsigned int &nb_coeur) {
    lseek(fd_cpuinfo, 0, SEEK_SET);
    freqs_sum = 0.0;
    nb_coeur  = 0;

    std::string str;
    while (readline(fd_cpuinfo, str)) {
        const char *c = str.c_str();
        if (strncmp(c, "cpu MHz", strlen("cpu MHz")) == 0) {
            freqs_sum += get_double_in_str(c);
            nb_coeur++;
        }
    }
}

/* extrait les 10 premiers entiers d'une ligne "cpu  ..." de /proc/stat */
void get_cpu_stat(unsigned long cpu_stat[10], const std::string &str) {
    const char *c_str = str.c_str();
    int space = 0;
    for (int i = 0; i < 10; i++) {
        int rel = get_indice_of_int(c_str + space);
        if (rel < 0) { cpu_stat[i] = 0; continue; }
        int ind = space + rel;
        cpu_stat[i] = strtoul(&c_str[ind], NULL, 10);
        int sp = get_indice_of_space(c_str + ind);
        if (sp < 0) break;
        space = ind + sp;
    }
}

/* récupère la ligne "cpu" de /proc/stat 2 fois avec un sleep(2) entre
   on lseek AVANT chacune des deux lectures (sinon à la 2e itération de
   la boucle main on est encore positionné à la fin du fichier) */
void parse_cpu_stats(int fd_stat, unsigned long cpu_stat[10], unsigned long cpu_stat2[10]) {
    std::string str;

    lseek(fd_stat, 0, SEEK_SET);
    readline(fd_stat, str);
    get_cpu_stat(cpu_stat, str);

    sleep(2);

    lseek(fd_stat, 0, SEEK_SET);
    readline(fd_stat, str);
    get_cpu_stat(cpu_stat2, str);
}

double get_cpu_idle_percentage(unsigned long cpu_stat[10], unsigned long cpu_stat2[10]) {
    unsigned long idle1 = cpu_stat[3]  + cpu_stat[4];
    unsigned long idle2 = cpu_stat2[3] + cpu_stat2[4];
    double delta_idle = (double)(idle2 - idle1);

    unsigned long total1 = 0, total2 = 0;
    for (int i = 0; i < 10; i++) {
        total1 += cpu_stat[i];
        total2 += cpu_stat2[i];
    }
    double delta_total = (double)(total2 - total1);
    if (delta_total <= 0) return 0.0;

    return (delta_idle / delta_total) * 100.0;
}

/* write() peut écrire moins que demandé : on boucle jusqu'à n octets */
ssize_t write_all(int fd, const void *buf, size_t n) {
    const char *p = (const char*)buf;
    size_t total = 0;
    while (total < n) {
        ssize_t w = write(fd, p + total, n - total);
        if (w < 0) {
            if (errno == EINTR) continue;
            return -1;
        }
        if (w == 0) return -1;
        total += (size_t)w;
    }
    return (ssize_t)total;
}

int main(int /*argc*/, char* /*argv*/[]) {
    int fd_meminfo = open("/proc/meminfo", O_RDONLY);
    int fd_cpuinfo = open("/proc/cpuinfo", O_RDONLY);
    int fd_stat    = open("/proc/stat",    O_RDONLY);
    if (fd_meminfo < 0 || fd_cpuinfo < 0 || fd_stat < 0) {
        std::cerr << "open /proc/... a echoue: " << strerror(errno) << std::endl;
        return 1;
    }

    uint64_t     mem_available = 0;    // en kB
    unsigned int nb_coeur      = 0;
    double       freqs_sum     = 0.0;
    unsigned long cpu_stat[10], cpu_stat2[10];

    // ouverture du socket Unix vers le programme Go
    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    if (sock < 0) {
        std::cerr << "socket() a echoue: " << strerror(errno) << std::endl;
        return 1;
    }

    sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, "/tmp/cpu.sock", sizeof(addr.sun_path) - 1);

    // retry tant que le programme Go n'a pas encore Listen sur /tmp/cpu.sock
    while (connect(sock, (sockaddr*)&addr, sizeof(addr)) < 0) {
        std::cerr << "connect /tmp/cpu.sock: " << strerror(errno)
                  << " -- retry dans 1s" << std::endl;
        sleep(1);
    }
    std::cout << "connecte a /tmp/cpu.sock" << std::endl;

    while (1) {
        get_mem_info(fd_meminfo, &mem_available);
        get_cpu_info(fd_cpuinfo, freqs_sum, nb_coeur);
        parse_cpu_stats(fd_stat, cpu_stat, cpu_stat2);   // contient sleep(2)

        double idle_pct      = get_cpu_idle_percentage(cpu_stat, cpu_stat2);
        double freqs_by_idle = freqs_sum * (idle_pct / 100.0);

        std::cout << "mem_available=" << mem_available
                  << " kB | coeurs=" << nb_coeur
                  << " | idle=" << idle_pct << "%"
                  << " | freqs*idle=" << freqs_by_idle << " MHz"
                  << std::endl;

        // envoi : uint64 (mem en kB) + double (freqs*idle en MHz) = 16 octets
        // le Go fait io.ReadFull sur 16 octets puis decode LittleEndian.
        // /!\ ce code suppose une machine little-endian (x86, ARM par defaut)
        if (write_all(sock, &mem_available, sizeof(mem_available)) < 0 ||
            write_all(sock, &freqs_by_idle, sizeof(freqs_by_idle)) < 0) {
            std::cerr << "write a echoue: " << strerror(errno) << ", sortie" << std::endl;
            break;
        }
    }

    close(sock);
    close(fd_cpuinfo);
    close(fd_meminfo);
    close(fd_stat);
    return 0;
}