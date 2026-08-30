Write this Go dev tool:

- When run, it watches the passed-in directory, which should be a Go project.
- It creates an AST-modified variant of the non-test code in a separate subfolder `__avoud_mirror`, mirroring the folder/file structure of the original code.
- On a `.go` file change, it updates the corresponding AST-modified file.

# AST-modified behavior

- Look for a custom comment pragma in format `fmt.Sprintf(`//%v:%v`, pragmaCode, directive)` before each func.
- Set `pragmaCode := "avoud"` for now.
- If pragma is `//avoud:pure` or `//avoud:deterministic`...
- Log such function names, with as fully qualified uniquely identifiable names as possible.
- Log inputted arguments and outputted returned values of such functions.
- Log all assignments within such functions.

# Example

From this original source code:
```
//avoud:pure
func Sum(a, b int) int {
    return a + b
}

//avoud:pure
func Format(txt string) string {
    upper := strings.ToUpper(txt)
    return strings.TrimSpace(upper)
}

func main() {
    sum := Sum(2, 3)
    name := Format(" homie ")
    fmt.Println(sum, name)
}
```

Generate a mirrored codebase that, when run, would log something like this example:
```
hint|pure|main.Sum
hint|pure|main.Format
input|main.Sum|2|3
output|main.Sum|5
input|main.Format|" homie "
assign|main.Format|upper|" HOMIE "
output|main.Format|"HOMIE"
```

- Log the output above in a terminal.
- Additionally, if CLI argument `--to_file` was specified, write this log to the path specified by `--output={DESIRED_PATH}` (user substitutes "{DESIRED_PATH}" with their path)
- If `--to_file` was specified but no `--output` path was specified, write this log to `avoud_events.txt` in the current folder.