package expressions

import (
	"regexp"
	"strings"
)

// Приоритет
func precedence(op byte) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	}
	return 0
}

// Объяснение работы Validate
// Функиця рекусрсивно читает каждую скобку и корневое выражение
// Оно проверяет структуру: {число или выражение в скобках(что внутри скобок не важно)} ({знак}{число или выражение в скобках})  // Вторая большая скобка со {знак} может повторяться любое колво раз
var mathRegRoot, _ = regexp.Compile(`^ *(\d+(\.\d+)?|\([^()]*((\([^()]*)*([^()]*\))*)[^()]*\))( *[+\-*/] *(\d+(\.\d+)?|\([^()]*((\([^()]*)*([^()]*\))*)[^()]*\)))* *$`) // 💀

func Validate(infix string) bool {
	if !mathRegRoot.MatchString(infix) || strings.Count(infix, "(") != strings.Count(infix, ")") {
		return false
	}
	res := ""
	brCount := 0
	brr := false
	for _, i := range infix {
		switch string(i) {
		case "(":
			brCount += 1
			brr = true
		case ")":
			brCount -= 1
			if brCount < 0 {
				return false
			}
		}
		if brCount != 0 {
			res += string(i)
		}
		if brCount == 0 && len(res) >= 2 && brr {
			if string(i) != infix && !Validate(res[1:]) {
				return false
			}
			res = ""
			brr = false
		}
	}
	return true
}

// InfixToPostfix Функция для преобразования инфиксной формы записи выражения в постфиксную
func InfixToPostfix(infix string) string {
	var result strings.Builder
	var stack []byte

	for _, ch := range infix {
		switch ch {
		case ' ':
			continue
		case '+', '-', '*', '/':
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(byte(ch)) {
				result.WriteByte(' ')
				result.WriteByte(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			result.WriteByte(' ')
			stack = append(stack, byte(ch))
		case '(':
			stack = append(stack, byte(ch))
		case ')':
			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				result.WriteByte(' ')
				result.WriteByte(stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
		default:
			// Если символ - цифра, добавляем ее в выходную строку
			result.WriteString(string(ch))
		}
	}

	for len(stack) > 0 {
		result.WriteByte(' ')
		result.WriteByte(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return result.String()
}
