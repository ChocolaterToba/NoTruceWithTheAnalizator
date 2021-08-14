How to use this???
1. Install Golang
2. Run following from this project's root directory (the one where main.go is):
   $go build -o pathtoyourlib/nameofyourlib.so -buildmode=c-shared main.go
   pathtoyourlib may be omitted (for example, ... -o mylibfile.so ...), project's root directory is assumed in that case
3. Notice how nameofyourlib.so and nameofyourlib.h appeared in "pathtoyourlib" directory.
   These are your shared library and said library's header files respectfully.
4. Use these files to your advantage, for example, re-compiling lib with gcc to change headers if needed.
