/^## Options/,/^### Config File/ {
    /^\(`\|#\)/!d           # Delete any line not beginning with ` or #
    /```sh/ {               # Shove options.txt.tmp in after ```sh
        r options.txt.tmp
    }
}