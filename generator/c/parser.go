package c

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Parse(path string, defines []string) (*Node, error) {
	// Setup command
	cmd := exec.Command("clang", "-fsyntax-only", "-fparse-all-comments", "-Xclang", "-ast-dump=json")

	for _, define := range defines {
		cmd.Args = append(cmd.Args, "-D"+define)
	}

	cmd.Args = append(cmd.Args, path)

	// Capture stderr
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run and parse json output
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	var root *Node
	if err := json.NewDecoder(out).Decode(&root); err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, stderr.String())
	}

	// Fix Node.Loc.File
	fixLocations(root.Inner, "")

	// Parse macro defines
	err = parseMacroDefines(path, defines, root)
	if err != nil {
		return nil, err
	}

	// Attach documentation comments to macro defines
	err = attachMacroComments(path, root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func fixLocations(nodes []*Node, lastFile string) string {
	for _, node := range nodes {
		if node.Loc.File == "" {
			node.Loc.File = lastFile
		} else {
			lastFile = node.Loc.File
		}

		lastFile = fixLocations(node.Inner, lastFile)
	}

	return lastFile
}

func parseMacroDefines(path string, defines []string, root *Node) error {
	// Setup command
	cmd := exec.Command("clang", "-E", "-dM")

	for _, define := range defines {
		cmd.Args = append(cmd.Args, "-D"+define)
	}

	cmd.Args = append(cmd.Args, path)

	// Capture stderr
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run the command
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%w:\n%s", err, stderr.String())
	}

	// Parse defines
	for line := range strings.Lines(string(out)) {
		line, ok := strings.CutPrefix(line, "#define ")
		if !ok {
			continue
		}

		space := strings.IndexRune(line, ' ')

		var name string
		var value string

		if space == -1 {
			name = line
		} else {
			name = line[:space]
			value = line[space+1:]
		}

		if strings.HasPrefix(name, "__") || strings.ContainsRune(name, '(') {
			continue
		}

		root.Inner = append(root.Inner, &Node{
			Kind:  "MacroDefine",
			Name:  strings.TrimSpace(name),
			Value: strings.TrimSpace(value),
		})
	}

	return nil
}

func attachMacroComments(path string, root *Node) error {
	macros := make(map[string]*Node)
	for _, n := range root.Inner {
		if n.Kind == "MacroDefine" {
			macros[n.Name] = n
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var commentBlock []string
	var groupCommentBlock []string

	inBlockComment := false
	inGroup := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		isComment := false

		// Handle block comments (/* ... */)
		if inBlockComment {
			commentBlock = append(commentBlock, trimmed)
			isComment = true
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
		} else if strings.HasPrefix(trimmed, "/*") {
			commentBlock = []string{trimmed}
			isComment = true
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
		} else if strings.HasPrefix(trimmed, "//") {
			commentBlock = append(commentBlock, trimmed)
			isComment = true
		}

		if isComment {
			// Check for Doxygen group start
			if strings.Contains(trimmed, "@{") {
				inGroup = true
				groupCommentBlock = append([]string(nil), commentBlock...)
				commentBlock = nil // Prevent using this as an item-specific comment!
			}

			// Check for Doxygen group end
			if strings.Contains(trimmed, "@}") {
				inGroup = false
				groupCommentBlock = nil
				commentBlock = nil // Prevent the closing marker from becoming a comment
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#define") {
			parts := strings.Fields(trimmed)

			if len(parts) >= 2 {
				name := parts[1]
				if idx := strings.IndexByte(name, '('); idx != -1 {
					name = name[:idx]
				}

				if node, ok := macros[name]; ok {
					// 1. Assign specific comment if one exists
					if len(commentBlock) > 0 {
						node.Text = cleanMacroComment(commentBlock)
					}
					// 2. Assign group comment if we are inside a group
					if inGroup && len(groupCommentBlock) > 0 {
						node.GroupText = cleanMacroComment(groupCommentBlock)
					}
				}
			}

			commentBlock = nil
			continue
		}

		commentBlock = nil
	}

	return nil
}

func cleanMacroComment(lines []string) string {
	var processed []string

	for _, line := range lines {
		// Strip outer block comment syntax
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "/*!")
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.TrimSpace(line)

		// Strip single line comment syntax or continuation stars
		if strings.HasPrefix(line, "*") {
			line = line[1:]
		} else if strings.HasPrefix(line, "///") {
			line = line[3:]
		} else if strings.HasPrefix(line, "//!") {
			line = line[3:]
		} else if strings.HasPrefix(line, "//") {
			line = line[2:]
		}

		// Strip Doxygen group markers to prevent them from bleeding into the output docs
		line = strings.ReplaceAll(line, "@{", "")
		line = strings.ReplaceAll(line, "@}", "")

		// Trim a single leading space if present (e.g. `* Hello` -> `Hello`)
		if strings.HasPrefix(line, " ") {
			line = line[1:]
		}

		// Trim trailing spaces
		line = strings.TrimRight(line, " \t")

		processed = append(processed, line)
	}

	// Remove leading empty lines
	for len(processed) > 0 && processed[0] == "" {
		processed = processed[1:]
	}

	// Remove trailing empty lines
	for len(processed) > 0 && processed[len(processed)-1] == "" {
		processed = processed[:len(processed)-1]
	}

	return strings.Join(processed, "\n")
}
