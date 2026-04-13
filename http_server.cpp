#include <cstdio>
#include <cstdlib>

#include <sys/stat.h>

#include "includes/httplib.h"

#define BUFSIZE 4096

long get_port(char *str) {
    char *end;
    long port = strtol(str, &end, 10); // on convertit la string en long

    if(*end != '\0') // on verifie que la string ne contenait que des chiffres
        return -1;

    return port;
}

int main(int argc, char *argv[]) {
    if(argc < 2) {
        std::cout << "Pas assez d'argument (port)" << std::endl;
        return 1;
    }

    // récupération du port donné en argument
    long port = get_port(argv[1]);
    if(port < 0) {
        std::cout << "Port mal formé" << std::endl;
        return 1;
    }

    // Serveur http
    httplib::Server svr;

    svr.Get("/", [](const httplib::Request&, httplib::Response& res) {
        res.set_content("Hello, World!", "text/plain");
    });

    svr.Post("/metrics", [](const auto &req, auto &res) {
        int fd = open("/proc/meminfo", O_RDONLY);
        int r;
        char *buf = (char *)malloc(sizeof(char) * BUFSIZE);
        while((r = read(fd, buf, BUFSIZE)) > 0)
            // récupérer données (info cpu + info mem)
            std::cout << buf;

        close(fd);

        std::string json = R"({"cpu":30, "mem":40})"; // info à modif
        res.set_content(json, "application/json");
    });

    svr.listen("127.0.0.1", port); // hébergé en localhost

    return 0;
}
