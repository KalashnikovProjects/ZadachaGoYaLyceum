package expressions

import (
	"errors"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/my_errors"
	"testing"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		infix    string
		expected error
	}{
		{"1 + 2 + 6", nil},
		{"1 + 2 + 3a", my_errors.ExpressionValidateError},
		{"1 + 2 ++ 6", my_errors.ExpressionValidateError},
		{"3 + 4 * (2 - 1) * 5", nil},
		{"(1+2)*3", nil},                                    // Верный ввод
		{"((1+2)*3)", nil},                                  // Верный ввод с двойными скобками
		{"((1+2)*3", my_errors.ExpressionValidateError},     // Неверный ввод, недостающая закрывающая скобка
		{"(1+2)*3)", my_errors.ExpressionValidateError},     // Неверный ввод, недостающая открывающая скобка
		{"(1+2)*(3", my_errors.ExpressionValidateError},     // Неверный ввод, недостающая закрывающая скобка
		{"(1+2)3", my_errors.ExpressionValidateError},       // Неверный ввод, отсутствие оператора между числом и выражением в скобках
		{"(1+2)*3*(", my_errors.ExpressionValidateError},    // Неверный ввод, лишняя открывающая скобка
		{"(1+(2*3", my_errors.ExpressionValidateError},      // Неверный ввод, недостающая закрывающая скобка
		{"(1+2)3*(4+5)", my_errors.ExpressionValidateError}, // Неверный ввод, отсутствие оператора между выражениями в скобках
		{"(1+2)*3*(4+5)", nil},                              // Верный ввод с несколькими уровнями скобок
		{"( 1 + 2 ) * ( 4 + 5 )", nil},                      // Верный ввод с несколькими уровнями скобок
		{"((1+2)*(4+5))*((6+7)/(8+9))", nil},                // Верный ввод с множеством уровней скобок
		{"1 + 2)", my_errors.ExpressionValidateError},       // Много закрывающих скобок
		//{"(1+2)(3+4)", my_errors.ExpressionValidateError},               // Неверный ввод, отсутствие оператора между выражениями в скобках
	}

	for _, tc := range testCases {
		if result := Validate(tc.infix); !errors.Is(result, tc.expected) {
			t.Errorf("Expected Validate(%q) to be %v, but got %v", tc.infix, tc.expected, result)
		}
	}
}

func TestValidateAdditionalCases(t *testing.T) {
	testCases := []struct {
		infix    string
		expected error
	}{
		{"a + b * c", my_errors.ExpressionValidateError}, // Буквенные переменные
		{"3.14 + 2.5", nil}, // Float числа
		{"(a + b) * (c - d)", my_errors.ExpressionValidateError},         // Буквенные переменные в скобках
		{"(3.14 + 2.5) * ((6.7 * (3 - 4) / 5) - 4.2)", nil},              // Float числа в скобках
		{"3 + 4 * (2 - 1) * 5", nil},                                     // Выражение без скобок
		{"3 + 4 * (2 - 1) * 5 +", my_errors.ExpressionValidateError},     // Неверный ввод, оператор в конце
		{"3 + 4 * (2 - 1) * 5 /", my_errors.ExpressionValidateError},     // Неверный ввод, оператор в конце
		{"3 + 4 * (2 - 1) * 5 *", my_errors.ExpressionValidateError},     // Неверный ввод, оператор в конце
		{"(3.14 + 2.5) * (6.7 - 4.2", my_errors.ExpressionValidateError}, // Неверный ввод, недостающая закрывающая скобка
		{"3.14 + 2.5) * (6.7 - 4.2)", my_errors.ExpressionValidateError}, // Неверный ввод, лишняя закрывающая скобка
		{"3 + * 4", my_errors.ExpressionValidateError},                   // Неверный ввод, лишний оператор
		{"3 + 4 * 2 -", my_errors.ExpressionValidateError},               // Неверный ввод, оператор в конце
	}

	for _, tc := range testCases {
		if result := Validate(tc.infix); !errors.Is(result, tc.expected) {
			t.Errorf("Expected Validate(%q) to be %v, but got %v", tc.infix, tc.expected, result)
		}
	}
}

func TestValidateLargeExpressions(t *testing.T) {
	testCases := []struct {
		infix    string
		expected error
	}{
		{"(1 + 2) * (3 + 4) / (5 + 6) - (7 + 8) * (9 + 10)", nil},                              // Сложное выражение с операциями +, -, *, /
		{"((2 * 3) - (4 / 2)) * ((8 / 2) + (6 - 2))", nil},                                     // Сложное выражение с операциями +, -, *, /
		{"((1 + 2) * (3 - 4) / (5 * 6) - (7 * 8) + (9 - 10))", nil},                            // Сложное выражение с операциями +, -, *, /
		{"(a + b) * (c - d) / (e * f) - (g + h) * (i + j)", my_errors.ExpressionValidateError}, // Сложное выражение с переменными и операциями +, -, *, /
		{"(1.2 + 3.4) * (5.6 - 7.8) / (9.1 * 2.3) - (4.5 + 6.7) * (8.9 + 0.1)", nil},           // Сложное выражение с числами с плавающей точкой и операциями +, -, *, /
		{"3 + 4 * (2 - 1) * 5", nil},                                                           // Простое выражение без скобок
		{"(1 + 2)3 * (4 + 5)", my_errors.ExpressionValidateError},                              // Неверный ввод, отсутствие оператора между числом и выражением в скобках
		{"(1 + 2) * 3 * (4 + 5)", nil},                                                         // Верный ввод с несколькими уровнями скобок
		{"((1 + 2) * (4 + 5)) * ((6 + 7) / (8 + 9))", nil},                                     // Сложное выражение с множеством уровней скобок
		{"(1 + 2)(3 + 4)", my_errors.ExpressionValidateError},                                  // Неверный ввод, отсутствие оператора между выражениями в скобках
		{"3 + 4 * 2 -", my_errors.ExpressionValidateError},                                     // Неверный ввод, оператор в конце
		{"3 + * 4", my_errors.ExpressionValidateError},                                         // Неверный ввод, лишний оператор
	}

	for _, tc := range testCases {
		if result := Validate(tc.infix); !errors.Is(result, tc.expected) {
			t.Errorf("Expected Validate(%q) to be %v, but got %v", tc.infix, tc.expected, result)
		}
	}
}

func TestValidateMoreExpressions(t *testing.T) {
	testCases := []struct {
		infix    string
		expected error
	}{
		{"(a + b) * c / (d - e) * (f + g)", my_errors.ExpressionValidateError}, // Выражение с переменными и операциями +, -, *, /
		{"1 + 2 * 3 - 4 / 5", nil},                                                 // Простое выражение с операциями +, -, *, /
		{"(1 + 2 * 3 - 4 / 5) * (6 + 7 - 8 * 9)", nil},                             // Большое выражение с операциями +, -, *, /
		{"((2 + 3) * 4 - 5) / (6 + 7) * 8", nil},                                   // Сложное выражение с операциями +, -, *, /
		{"(2 + 3) * (4 - 5) / (6 + 7) * 8", nil},                                   // Сложное выражение с операциями +, -, *, /
		{"(1 + 2) * (3 + 4) * (5 + 6)", nil},                                       // Большое выражение с операциями +, *, ()
		{"(1 + 2) * 3 + 4", nil},                                                   // Выражение с операциями +, *
		{"1 + 2 * 3 * 4", nil},                                                     // Выражение с операциями +, *
		{"(1 + 2) * 3 / (4 + 5)", nil},                                             // Выражение с операциями +, *, /
		{"1 * 2 / (3 + 4) - 5", nil},                                               // Выражение с операциями *, /, -
		{"((1 + 2) * 3 - 4) / 5", nil},                                             // Выражение с операциями +, -, *, /
		{"((1 + 2) * 3 - 4) / (5 + 6)", nil},                                       // Сложное выражение с операциями +, -, *, /
		{"((1 + 2) * 3 - 4) / (5 + 6) * 7", nil},                                   // Сложное выражение с операциями +, -, *, /
		{"((1 + 2) * 3 - 4) / (5 + 6) * (7 - 8)", nil},                             // Сложное выражение с операциями +, -, *, /
		{"((1 + 2) * 3 - 4) / (5 + 6) * (7 - 8) / (9 + 10)", nil},                  // Сложное выражение с операциями +, -, *, /
		{"((1 + 2) * 3 - 4) / (5 + 6) * (7 - 8) / (9 + 10) - 11 * (12 - 13)", nil}, // Сложное выражение с операциями +, -, *, /
	}

	for _, tc := range testCases {
		if result := Validate(tc.infix); !errors.Is(result, tc.expected) {
			t.Errorf("Expected Validate(%q) to be %v, but got %v", tc.infix, tc.expected, result)
		}
	}
}

func TestInfixToPostfix(t *testing.T) {
	testCases := []struct {
		infix    string
		expected string
	}{
		{"1 + 2", "1 2 +"},
		{"1 + 2 * 3", "1 2 3 * +"},
		{"(1 + 2) * 3", "1 2 + 3 *"},
		{"3 + 4 * (2 - 1) * 5", "3 4 2 1 - * 5 * +"},
		{"(1+2)*(3+4)/(5+6)", "1 2 + 3 4 + * 5 6 + /"},
		{"a + b * c", "a b c * +"},
		{"3 + (4 * 2 - 1) * 5", "3 4 2 * 1 - 5 * +"},
		{"1 + 2 * (3 - 4 / 5)", "1 2 3 4 5 / - * +"},
	}

	for _, tc := range testCases {
		result := InfixToPostfix(tc.infix)
		if result != tc.expected {
			t.Errorf("Expected InfixToPostfix(%q) to be %q, but got %q", tc.infix, tc.expected, result)
		}
	}
}
