#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>

int main() {
    int64_t result = ((10 + 5) * 2);
    printf("%lld\n", (long long)(result));
    if ((result == 30)) {
        printf("%lld\n", (long long)(1));
    } else {
        printf("%lld\n", (long long)(0));
    };
    return 0;
}
