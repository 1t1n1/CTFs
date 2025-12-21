#!/usr/bin/env bash


DB_PATH=${DB_PATH:-/tmp/fsdb.tmp}

function init_db() {
    mkdir -p $DB_PATH
    mkdir -p $DB_PATH/keys
    mkdir -p $DB_PATH/values
}

function resp_ok() {
    echo "OK"
}

function resp_err() {
    echo "ERR"
}

function resp_check() {
    if [[ $1 -eq 0 ]]; then
        resp_ok
    else
        resp_err
    fi
}

function key_path() {
    key=$1
    # remove last directory separator
    key2=${key%/}
    mkdir -p $(dirname "$DB_PATH/keys/$key2")
    # store original key in keys directory
    echo "$key" > "$DB_PATH/keys/$key"
    # store value in values directory
    mkdir -p $(dirname "$DB_PATH/values/$key2")
    echo "$DB_PATH/values/$key"
}

function get_value() {
    if [[ ${1:0:1} == '$' ]]; then
        var=${1:1}
        echo ${!var:-ERR: variable `$var` not defined}
    else
        echo "$1"
    fi
}

function cmd_status() {
    resp_ok
}

function cmd_exit() {
    resp_ok
    exit 0
}

function cmd_get() {
    key=$(key_path $1)
    if [[ -f $key ]]; then
        cat $key | wc -c
        cat $key
        echo
    else
        resp_err
    fi
}

function cmd_set() {
    echo -n $(get_value $2) > $(key_path $1)
    resp_check $?

}

function cmd_rename() {
    mv $(key_path $1) $(key_path $2)
    resp_check $?
}

function cmd_del() {
    rm -rf $(key_path $1)
    resp_check $?
}

function cmd_flushdb() {
    rm -rf "$DB_PATH/keys"
    rm -rf "$DB_PATH/values"
    init_db
    resp_ok
}

function cmd_list() {
    ls $DB_PATH/keys | wc -l

    for file in $(ls $DB_PATH/keys); do
        data=$(cat "$DB_PATH/keys/$file")
        printf $data | wc -m
        printf $data 
        printf "\n"
    done
}


function eval_query() {
    index=1
    while IFS= read -r line
    do
        echo -n "$index: "
        line="cmd_$line"
        eval $line
        index=$((index+1))
    done
}


function main() {
    init_db
    cat "$1" | eval_query
}

main "$1"
