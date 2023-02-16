package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Stack ...
type Stack struct {
	data [1024]string
	top  int
}

// e.g. 5 ∈ [ 3 , 8 ]
var scopeExp = "([1-9]\\d*)\\s*∈\\s*\\[\\s*([1-9]\\d*)\\s*,\\s*([1-9]\\d*)\\s*\\]"

//	Calculate Four simple operations：Support +, -, *, / and ^ (power)
//	input : "9+(3-1)*3+10/2"  output: 20.
func Calculate(express string) (int64, error) {
	if len(express) == 0 {
		return 0, fmt.Errorf("express is null")
	}
	express = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(express, " ", ""), "\n", ""), "\r", ""), "\t", "")
	suffixExpress, err := convertToSuffixExpress(express)
	if err != nil {
		return 0, err
	}
	return calc(suffixExpress)
}
func priority(s byte) int {
	switch s {
	case '+':
		return 1
	case '-':
		return 1
	case '*':
		return 2
	case '/':
		return 2
	case '^':
		return 3
	}
	return 0
}

//	convertToSuffixExpress Convert infix expression to suffix expression
// input: "9+(3-1)*3+10/2" output: [9 3 1 - 3 * + 10 2 / +]
func convertToSuffixExpress(express string) (suffixExpress []string, err error) {
	var (
		opStack Stack // Operator stack
		i       int
	)
LOOP:
	// step1: Scan infix expression from left to right
	for i < len(express) {
		switch {
		//	1. If the operand is read, the operand is stored in the suffix expression.
		case express[i] >= '0' && express[i] <= '9':
			var number []byte // For example, the number 123 consists of '1', '2' and '3'
			for ; i < len(express); i++ {
				if express[i] < '0' || express[i] > '9' {
					break
				}
				number = append(number, express[i])
			}
			suffixExpress = append(suffixExpress, string(number))
		//	2. If operator is read：
		//	"(" case: Then it is pushed directly into the operator stack。
		case express[i] == '(':
			opStack.Push(fmt.Sprintf("%c", express[i]))
			i++
		//	")" case: The suffix expression in the operator stack is output until the left parenthesis is encountered.
		case express[i] == ')':
			for !opStack.IsEmpty() {
				data, _ := opStack.Pop()
				if data[0] == '(' {
					break
				}
				suffixExpress = append(suffixExpress, data)
			}
			i++
		//	supported operator: +, -, *, /, ^(power)
		case express[i] == '+' || express[i] == '-' || express[i] == '*' || express[i] == '/' || express[i] == '^':
			// a. If the operator stack is empty, it is pushed directly into the operator stack.
			if opStack.IsEmpty() {
				opStack.Push(fmt.Sprintf("%c", express[i]))
				i++
				continue LOOP
			}
			data, _ := opStack.Top()
			//	b. If the operator at the top of the operator stack is a bracket, it is directly pushed into the operator stack. (it can only be left parenthesis)
			if data[0] == '(' {
				opStack.Push(fmt.Sprintf("%c", express[i]))
				i++
				continue LOOP
			}
			//	c. If the priority is lower or equal than the operator at the top of the operator stack, the operator at the top of the stack is output to the suffix expression until the stack is empty or a higher priority than the current operator is found. And pushes the current operator into the operator stack.
			if priority(express[i]) <= priority(data[0]) {
				tmp := priority(express[i])
				for !opStack.IsEmpty() && tmp <= priority(data[0]) {
					suffixExpress = append(suffixExpress, data)
					opStack.Pop()
					data, _ = opStack.Top()
				}
				opStack.Push(fmt.Sprintf("%c", express[i]))
				i++
				continue LOOP
			}
			//	d. If the priority is higher than the operator at the top of the operator stack, it is directly pushed into the operator stack.
			opStack.Push(fmt.Sprintf("%c", express[i]))
			i++
		default:
			err = fmt.Errorf("invalid express:%s", string(express[i]))
			return
		}
	}
	//	step2. Pop up the operators in the operator stack in turn and store them in the suffix expression.
	for !opStack.IsEmpty() {
		data, _ := opStack.Pop()
		if data[0] == '#' {
			break
		}
		suffixExpress = append(suffixExpress, data)
	}
	return
}

//	calc calculate the suffix expressions
func calc(suffixExpress []string) (result int64, err error) {
	var (
		num1    string
		num2    string
		opStack Stack // Operation stack, used to store operands and operators
	)
	// step1: Scan infix expression from left to right
	for i := 0; i < len(suffixExpress); i++ {
		var cur = suffixExpress[i]
		//	1. If the operator is read
		if cur[0] == '+' || cur[0] == '-' || cur[0] == '*' || cur[0] == '/' || cur[0] == '^' {
			//	Pop up two data from the operation stack for operation
			num1, err = opStack.Pop()
			if err != nil {
				return
			}
			num2, err = opStack.Pop()
			if err != nil {
				return
			}
			//	The first ejected data is B, and the later ejected data is A
			B, _ := strconv.ParseFloat(num1, 64)
			A, _ := strconv.ParseFloat(num2, 64)
			var res float64
			switch cur[0] {
			case '+':
				res = A + B
			case '-':
				res = A - B
			case '*':
				res = A * B
			case '/':
				res = A / B
			case '^':
				res = math.Pow(A, B)
			default:
				err = fmt.Errorf("invalid operation")
				return
			}
			//	push middle result
			opStack.Push(fmt.Sprintf("%.6f", res))
		} else {
			//	If the operand is read, press the stack directly
			opStack.Push(cur)
		}
	}
	//	After the calculation, the final result is saved at the top of the stack
	resultStr, err := opStack.Top()
	if err != nil {
		return
	}

	floatRes, err := strconv.ParseFloat(resultStr, 64)
	result = int64(floatRes)
	return
}

// IsEmpty ...
func (s *Stack) IsEmpty() bool {
	return s.top == 0
}

// Top ...
func (s *Stack) Top() (ret string, err error) {
	if s.top == 0 {
		err = fmt.Errorf("stack is empty")
		return
	}
	ret = s.data[s.top-1]
	return
}

// Push ...
func (s *Stack) Push(str string) {
	s.data[s.top] = str
	s.top++
}

// Pop ...
func (s *Stack) Pop() (ret string, err error) {
	if s.top == 0 {
		err = fmt.Errorf("stack is empty")
		return
	}
	s.top--
	ret = s.data[s.top]
	return
}

// CalculateCondExp Calculate result of conditional expression
func CalculateCondExp(expression string) bool {
	switch {
	case strings.Contains(expression, "|"):
		var orResult bool
		orExps := strings.Split(expression, "|")
		for _, exp := range orExps {
			if strings.Contains(exp, "&") {
				orResult = orResult || getANDResult(exp)
				continue
			}

			if orResult {
				return orResult
			}

			orResult = orResult || getSingleCondResult(exp)
		}

		return orResult
	case strings.Contains(expression, "&"):
		return getANDResult(expression)
	default:
		return getSingleCondResult(expression)
	}
}

func getANDResult(expression string) bool {
	andResult := true
	andExps := strings.Split(strings.TrimSpace(expression), "&")
	for _, exp := range andExps {
		andResult = andResult && getSingleCondResult(exp)
		if !andResult {
			return andResult
		}
	}

	return andResult
}

func getSingleCondResult(condition string) bool {
	trimSpaceCond := strings.TrimSpace(condition)
	switch {
	case strings.Contains(trimSpaceCond, ">="):
		compares := strings.Split(trimSpaceCond, ">=")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf >= rt
		}
	case strings.Contains(trimSpaceCond, ">"):
		compares := strings.Split(trimSpaceCond, ">")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf > rt
		}
	case strings.Contains(trimSpaceCond, "<="):
		compares := strings.Split(trimSpaceCond, "<=")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf <= rt
		}

	case strings.Contains(trimSpaceCond, "<"):
		compares := strings.Split(trimSpaceCond, "<")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf < rt
		}

	case strings.Contains(trimSpaceCond, "!="):
		compares := strings.Split(trimSpaceCond, "!=")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf != rt
		}
	case strings.Contains(trimSpaceCond, "="):
		compares := strings.Split(trimSpaceCond, "=")
		if len(compares) == 2 {
			lf := strings.TrimSpace(compares[0])
			rt := strings.TrimSpace(compares[1])

			return lf == rt
		}

	case strings.Contains(trimSpaceCond, "∈"):
		return scopeCondResult(trimSpaceCond)
	}

	return false
}

func scopeCondResult(matchStr string) bool {
	reg := regexp.MustCompile(scopeExp)
	if reg == nil {
		return false
	}

	if reg.MatchString(matchStr) {
		parts := strings.Split(matchStr, "∈")
		if len(parts) != 2 {
			return false
		}

		compared, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)

		trimStr := strings.Trim(strings.Trim(strings.TrimSpace(parts[1]), "["), "]")
		rangeNos := strings.Split(trimStr, ",")

		lf, _ := strconv.ParseInt(strings.TrimSpace(rangeNos[0]), 10, 64)

		rt, _ := strconv.ParseInt(strings.TrimSpace(rangeNos[1]), 10, 64)

		if lf > rt {
			lf, rt = rt, lf
		}

		return compared >= lf && compared <= rt
	}

	return false
}

