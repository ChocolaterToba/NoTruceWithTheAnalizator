#include "stdio.h"
#include "lib/whatever.h"

int main() {
    char* result = LedChange(true, 1, true);
    printf("%s\n", result);
    return 0;
}