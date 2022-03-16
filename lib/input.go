package lib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("文件打开失败: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var targets []string

	br := bufio.NewReader(file)

	for {
		target, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}

		targets = append(targets, string(target))
	}

	if len(targets) == 0 {
		err := errors.New(fmt.Sprintf("输入文件为空！\n"))
		return nil, err
	}

	return targets, nil
}
