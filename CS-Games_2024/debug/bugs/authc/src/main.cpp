#include <stdio.h>
#include <string.h>
#include <string>

#include "crc.h"

#define CMD_LOGIN 0b0001
#define CMD_ADMIN 0b0010
#define CMD_2FA   0b0100
#define CMD_ERROR -1

size_t readCommand() {
  char command[8];
  scanf("%s", command);

  if(strcmp(command, "login") == 0)
    return CMD_LOGIN;

  if(strcmp(command, "2fa") == 0)
    return readCommand() | CMD_2FA;


  if(strcmp(command, "admin") == 0)
    return readCommand() | CMD_ADMIN;


  return CMD_ERROR;
}

#define EQ(a, b) a == b
#define HAS_FLAG(cmd, flag) EQ((cmd & flag), flag)

typedef struct {
  char *name;
  char *password;
  char *seed_2fa;
  char required_2fa;
} User;

// define users
User users[] = {
  {"guest", "guest", "6598438549", 0},
  {"admin", "admin", "5943923439", 1},
  {"superadmin", "superadmin", "184932039", 1}
};


size_t handleCommand(size_t cmd) {
  User *user = 0;

  if(HAS_FLAG(cmd, CMD_LOGIN)) {
    printf("Enter username:\n");
    char username[8];
    scanf("%s", username);

    printf("Enter password:\n");
    char password[8];
    scanf("%s", password);

    for(size_t i = 0; i < sizeof(users) / sizeof(User); i++) {
      if(strcmp(users[i].name, username) == 0 && strcmp(users[i].password, password) == 0) {
        user = &users[i];
        break;
      }
    }

    if(!user) {
      printf("ERR: invalid username or password\n");
      return 1;
    }

    if(strstr(user->name,"admin") != 0) {
      if(HAS_FLAG(cmd, CMD_ADMIN) == 0) {
        printf("ERR: admin access requires admin privileges\n");
        return 1;
      }
    }

      if(HAS_FLAG(cmd, CMD_2FA)) {
      printf("Current time:\n");
      int time;
      scanf("%d", &time);

      printf("Enter 2fa code:\n");
      int code;
      scanf("%d", &code);

      char buf[32];
      sprintf(buf, "%s-%d", user->seed_2fa, time);

      int expected = crc32(0, buf, strlen(buf)) % 1000000;

      if(expected != code) {
        printf("ERR: invalid 2fa code\n");
        printf("expected: %d\n", expected);
        return 1;
      }

    } else {
        if(user->required_2fa) {
          printf("ERR: 2fa required\n");
          return 1;
        }
    }

    printf("OK: access granted for %s\n", user->name);
  }

  return 0;
}

int main(int argc, char *argv[]) {
  size_t cmd = readCommand();

  return handleCommand(cmd);
}
