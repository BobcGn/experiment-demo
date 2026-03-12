package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// 求和（使用切片避免数组复制）
func sigma(arr []float64) float64 {
	var result = float64(0)
	for i := 0; i < len(arr); i++ {
		result += arr[i]
	}
	return result
}

// 计算平均值（使用切片避免数组复制）
func average(arr []float64) float64 {
	return sigma(arr) / float64(len(arr))
}

// 计算标准偏差（自动计算平均值，使用切片避免数组复制）
func S1(li []float64) float64 {
	n := len(li)
	if n < 2 {
		return 0
	}
	// 先计算平均值
	L_avg := average(li)

	// 计算每个数据与平均值的差的平方和
	var sumSquares float64
	for i := 0; i < n; i++ {
		diff := li[i] - L_avg
		sumSquares += diff * diff
	}
	// 样本标准偏差公式: sqrt(Σ(xi - x̄)² / (n-1))
	return math.Sqrt(sumSquares / float64(n-1))
}

// A类不确定度（自动计算标准偏差，使用切片避免数组复制）
func Uncertainty_A(li []float64) float64 {
	n := len(li)
	return S1(li) / math.Sqrt(float64(n))
}

// B类不确定度
func Uncertainty_B(delta float64) float64 {
	return delta / 2 / math.Sqrt(3)
}

// 合成不确定度
func Uncertainty合成(A, B float64) float64 {
	result := A*A + B*B
	return math.Sqrt(result)
}

func main() {
	fmt.Println("=== 物理实验不确定度计算程序 ===")
	fmt.Println("支持计算：标准偏差、A类不确定度、B类不确定度、合成不确定度")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// 输入实验数据组数
	var groupCount int
	for {
		fmt.Print("请输入实验数据的组数：")
		fmt.Print("(例如：3 表示有3组数据)\n")
		fmt.Print("请输入： ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		count, err := strconv.Atoi(input)
		if err != nil || count <= 0 {
			fmt.Printf("错误：请输入一个正整数\n")
			continue
		}
		groupCount = count
		break
	}

	// 输入每组的测量次数
	var measureCount int
	for {
		fmt.Print("\n请输入每组的测量次数：")
		fmt.Print("(例如：6 表示每组6次测量)\n")
		fmt.Print("请输入： ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		count, err := strconv.Atoi(input)
		if err != nil || count <= 0 {
			fmt.Printf("错误：请输入一个正整数\n")
			continue
		}
		measureCount = count
		break
	}

	// 输入仪器误差限（B类不确定度参数）
	var delta float64
	for {
		fmt.Println("\n请输入仪器的误差限 (delta) ：")
		fmt.Print("（例如：0.01\n）")
		fmt.Print("请输入： ")

		deltaInput, _ := reader.ReadString('\n')
		deltaInput = strings.TrimSpace(deltaInput)
		deltaVal, err := strconv.ParseFloat(deltaInput, 64)
		if err != nil {
			fmt.Printf("错误：无法解析delta值 '%s'\n", deltaInput)
			continue
		}
		delta = deltaVal
		break
	}

	// 存储所有组的结果
	type Result struct {
		GroupNum int
		Data     []float64
		Avg      float64
		StdDev   float64
		U_A      float64
		U_B      float64
		U        float64
	}
	results := make([]Result, 0, groupCount)

	// 循环接收每组数据
	for group := 1; group <= groupCount; group++ {
		fmt.Printf("\n=== 第 %d 组数据 ===\n", group)
		fmt.Printf("请输入 %d 次测量的数据（用空格分隔）：\n", measureCount)
		fmt.Print("(例如：1.23 1.25 1.24 1.26 1.23 1.25)\n")
		fmt.Print("请输入： ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// 解析输入
		parts := strings.Fields(input)
		if len(parts) < measureCount {
			fmt.Printf("错误：需要输入%d个测量值\n", measureCount)
			group-- // 重新输入当前组
			continue
		}

		measurements := make([]float64, measureCount)
		valid := true
		for i := 0; i < measureCount; i++ {
			val, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				fmt.Printf("错误：无法解析第%d个值 '%s'\n", i+1, parts[i])
				valid = false
				break
			}
			measurements[i] = val
		}
		if !valid {
			group-- // 重新输入当前组
			continue
		}

		// 计算各项参数
		L_avg := average(measurements)
		s := S1(measurements)
		U_A := Uncertainty_A(measurements)
		U_B := Uncertainty_B(delta)
		U := Uncertainty合成(U_A, U_B)

		// 保存结果
		result := Result{
			GroupNum: group,
			Data:     measurements,
			Avg:      L_avg,
			StdDev:   s,
			U_A:      U_A,
			U_B:      U_B,
			U:        U,
		}
		results = append(results, result)

		// 显示当前组的计算结果
		fmt.Printf("\n第 %d 组计算结果：\n", group)
		fmt.Printf("  测量数据：")
		for _, val := range measurements {
			fmt.Printf("%.4f ", val)
		}
		fmt.Println()
		fmt.Printf("  平均值 L：%.6f\n", L_avg)
		fmt.Printf("  标准偏差 s：%.6f\n", s)
		fmt.Printf("  A类不确定度 U_A：%.6f\n", U_A)
		fmt.Printf("  B类不确定度 U_B：%.6f\n", U_B)
		fmt.Printf("  合成不确定度 U：%.6f\n", U)
		fmt.Printf("  最终测量结果：L = %.6f ± %.6f\n", L_avg, U)
	}

	// 输出所有组的汇总结果
	fmt.Println("\n=== 所有组汇总结果 ===")
	for _, result := range results {
		fmt.Printf("第 %d 组：L = %.6f ± %.6f\n", result.GroupNum, result.Avg, result.U)
	}

	fmt.Println("\n程序结束，感谢使用！")
}
