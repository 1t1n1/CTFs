#include <stdio.h>
#include <string.h>

void backhere(char *arg)
{        
	char buffer[128];
	strcpy(buffer, arg);
}

int main(int argc, char **argv)
{
	backhere(argv[1]);
	printf("Your data is %s\n", argv[1]);
	return 0;
}
