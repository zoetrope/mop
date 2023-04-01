package core

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// UploadResult uploads the content as a formatted comment to the given issue.
// The content will be read from the specified file with the given offset.
// If removeEscSequences is true, it will remove ANSI escape sequences from the content.
// Returns the sum of the offset and the length of the content read
func UploadResult(client *Client, issue int, filepath string, offset int64, removeEscSequences bool) (int64, error) {
	content, err := readContent(filepath, offset)
	if err != nil {
		return 0, err
	}
	bytesRead := int64(len(content))

	if removeEscSequences {
		content = removeANSIEscapeSequences(content)
		content = removeBackspace(content)
	}
	content = convertNewline(content)

	comment := formatAsCodeBlock(content)
	err = client.PostComment(issue, comment)
	if err != nil {
		return 0, err
	}
	return offset + bytesRead, nil
}

func readContent(filepath string, offset int64) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := checkFileSize(file, offset); err != nil {
		return "", err
	}

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func checkFileSize(file *os.File, offset int64) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// The max size of a Issue comment is 65536 characters
	if (fileInfo.Size() - offset) > 65000 {
		return fmt.Errorf("file is too large")
	}

	return nil
}

func formatAsCodeBlock(content string) string {
	return "```\n" + content + "\n```\n"
}

func removeANSIEscapeSequences(input string) string {
	ansiEscapeRegex := regexp.MustCompile(`\x1B\[[0-?]*[ -/]*[@-~]`)
	return ansiEscapeRegex.ReplaceAllString(input, "")
}

func convertNewline(content string) string {
	return strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
		"\n", "\n",
	).Replace(content)
}

func removeBackspace(content string) string {
	return strings.ReplaceAll(content, "\b", "")
}
