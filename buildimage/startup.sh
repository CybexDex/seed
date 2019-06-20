#!/bin/bash

printf "
set timeout -1 
spawn ./seed --port=$1
expect {
\"Please input passowrd\" { send \"1\\\r\" }
}
expect eof
" | expect
