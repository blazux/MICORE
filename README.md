# MICORE
MInimalistic COmmand REpeater

## Start the server :
`micore_server` 

make it a service if you want/need to.

## Use the client :
- add a repeated command *(set repeat_count to -1 for infinite)*:
```micore add "<linux_command> [options]..." "<repeat_interval>" <repeat_count>```


- list commands being repeated :
```micore list```

- stop a repeat task :
```micore stop <task_id>```


## Example :
```
# micore add "LANG=C date && echo 'hello world !'" "10s" 3

# Output: Wed May 22 21:15:55 AST 2024
hello world !

Output: Wed May 22 21:16:05 AST 2024
hello world !

Output: Wed May 22 21:16:15 AST 2024
hello world !

```