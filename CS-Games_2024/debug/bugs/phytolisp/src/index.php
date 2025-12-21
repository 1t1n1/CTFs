<?php

// limit execution time to 1 second
set_time_limit(1);

// read input from file
$input = file_get_contents($argv[1]);

$prelude = file_get_contents(__DIR__ . "/prelude.lisp");

// lex input into tokens
$tokens = lex($prelude . $input);

// parse tokens into AST and evaluate
runProgram(parse($tokens));

/** LEXER
 * 1. split input into tokens
 * 2. remove whitespace
 * 3. return tokens
 */
function lex($str) {
    $tokens = [];
    $str = preg_replace("/;.*/", "", $str);
    $str = str_replace(["(", ")"], [" ( ", " ) "], $str);
    $str = preg_replace("/\s+/", " ", $str);

    foreach (explode(" ", $str) as $token) {
        if ($token != "") {

            if(is_numeric($token)) {
                $tokens[] = (int)$token;
            } else {
                $tokens[] = $token;
            }
        }
    }

    return $tokens;
}

/** PARSER
 * 1. parse tokens into an AST
 * 2. return AST
 */
function parse($tokens) {
    $i = 0;
    $program = [];
    while(true) {
        $call = parseCall($tokens, $i);
        if($call == null) {
            break;
        }

        $program[] = $call;
    }
    return $program;
}

/** EVALUATOR
 * 1. evaluate AST
 * 2. return result
 */
function parseCall($token, &$i = 0) {
    if ($i < count($token) && $token[$i] == "(") {
        $i++;
        $expr = [];

        while ($i < count($token) && $token[$i] != ")" && $token[$i] != null) {
            $expr[] = parseCall($token, $i);
        }

        $i++;
        return $expr;
    } else {
        if ($i < count($token)) {
            return $token[$i++];
        }
    }
}

/** EVALUATOR
 * 1. evaluate AST
 * 2. return result
 */
function evaluate($expr, &$env) {
    if (!is_array($expr)) {
        if (isset($env["$expr"])) {
            return $env[$expr];
        } else {
            return $expr;
        }
    }

    if (empty($expr)) {
        return;
    }
    $op = $expr[0];
    print($op . " ");
    $args = array_slice($expr, 1);

    if (is_string($op) && $op[0] == ":") {
        $op = $env[$op];
    }

    switch ($op) {
        case "lambda":
            $params = $args[0];
            $body = array_slice($args, 1);

            return function () use ($params, $body, &$env) {
                $args = func_get_args();
                $newEnv = $env;

                foreach ($params as $i => $param) {
                    $newEnv[":".$param] = $args[$i];
                }

                foreach ($body as $expr) {
                    $result = evaluate($expr, $newEnv);
                }

                return $result;
            };
        case "puts":
            foreach ($args as $arg) {
                $out =  evaluate($arg, $env);
                if (is_array($out)) {
                    echo json_encode($out);
                } else {
                    echo $out;
                }
                echo " ";
            }
            echo "\n";
            break;
        case "set":
            $val =  ($env[":".evaluate($args[0], $env)] = evaluate($args[1], $env));
            return $val;
            break;
        case "var":
            return $env[":".evaluate($args[0], $env)];
        case "head":
            return evaluate($args[0], $env)[0];
        case "cons":
            return array_merge([evaluate($args[0], $env)], [evaluate($args[1][2], $env)][0]);
        case "tail":
            return array_slice(evaluate($args[0], $env), 1);
        case "quote":
            return $args;
        case "+":
            $a = evaluate($args[0], $env);
            $b = evaluate($args[1], $env);
            return $a + $b;
        case "-":
            return evaluate($args[0], $env) - evaluate($args[1], $env);
        case "*":
            return evaluate($args[0], $env) * evaluate($args[1], $env);
        case "/":
            return evaluate($args[0], $env) / evaluate($args[1], $env);
        case "<":
            return evaluate($args[0], $env) < evaluate($args[1], $env);
        case ">":
            return evaluate($args[0], $env) > evaluate($args[1], $env);
        case "<=":
            return evaluate($args[0], $env) <= evaluate($args[1], $env);
        case ">=":
            return evaluate($args[0], $env) >= evaluate($args[1], $env);
        case "==":
            return evaluate($args[0], $env) == evaluate($args[1], $env);
        case "!=":
            return evaluate($args[0], $env) != evaluate($args[1], $env);
        case "len":
            return count(evaluate($args[0], $env));
        case "if":
            if (evaluate($args[0], $env)) {
                return evaluate($args[1], $env);
            } else {
                return evaluate($args[2], $env);
            }
        case "while":
            $i = 0;
            $result = null;
            while (evaluate($args[0], $env) && $i < 100) {
                foreach (array_slice($args, 1) as $expr) {
                    $result = evaluate($expr, $env);
                }

                $i++;
            }
            return $result;
        case "begin":
            $result = null;
            foreach ($args as $expr) {
                $result = evaluate($expr, $env);
            }
            return $result;
        default:
            if(is_callable($op)) {
                $args = array_map(function($arg) use ($env) {
                    return evaluate($arg, $env);
                }, $args);
                return call_user_func_array($op, $args);
            }
            if(is_array($op)) {
                $func = evaluate($op, $env);
                $args = array_map(function($arg) use ($env) {
                    return evaluate($arg, $env);
                }, $args);
                return call_user_func_array($func, $args);
            }
    }
}

function runProgram($program) {
    $env = [];

    foreach ($program as $expr) {
        evaluate($expr, $env);
    }
}
