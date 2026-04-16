#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <signal.h>
#include <sys/wait.h>
#include <stdbool.h>

// Fonction pour exécuter une commande externe

//@return:
// Valeur 2 signifie que le fils a été tué par un signal
// Valeur 1 indique une erreur de terminaison ou que le fork a échoué
// Valeur autre est la valeur que la fils a renvoyé

int executeCommandeExterne(char *path, char **argv)
{
    int value = 0;

    pid_t child_pid = fork();

    if (child_pid > 0) // Processus parent
    {
        int status;
        waitpid(child_pid, &status, 0); // Attente de la fin du processus fils

        // Vérification de l'état du processus fils
        if (WIFEXITED(status)) // Le fils s'est terminé normalement
        {
            value = WEXITSTATUS(status);
        }
        else if (WIFSIGNALED(status)) // Le fils a été tué par un signal
        {
            value = 2;
        }
        else // Le fils a échoué d'une manière non prévue
        {
            value = 1;
        }
    }
    else if (child_pid == 0) // Processus fils
    {
        // Exécution de la commande externe avec execvp
        if (execvp(path, argv) == -1)
        {
            perror("redirect_exec");
            exit(1); // Quitte le processus fils en cas d'échec
        }
    }
    else // Échec de la création du processus fils
    {
        value = 1;
        fprintf(stderr, "Échec du fork\n");
    }

    return value;
}
