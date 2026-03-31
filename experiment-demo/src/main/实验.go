package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
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

// Result 存储每组数据的计算结果
type Result struct {
	GroupNum int
	Data     []float64
	Avg      float64
	StdDev   float64
	U_A      float64
	U_B      float64
	U        float64
}

// 导出结果为CSV文件
func exportToCSV(results []Result, delta float64) (string, error) {
	// 创建exports目录
	exportDir := "./exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", err
	}

	// 生成文件名
	fileName := fmt.Sprintf("实验数据结果_%s.csv", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(exportDir, fileName)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"组号", "测量数据", "平均值", "标准偏差", "A类不确定度", "B类不确定度", "合成不确定度", "最终结果"}
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// 写入数据
	for _, result := range results {
		dataStr := ""
		for i, val := range result.Data {
			if i > 0 {
				dataStr += " "
			}
			dataStr += fmt.Sprintf("%.4f", val)
		}

		row := []string{
			fmt.Sprintf("%d", result.GroupNum),
			dataStr,
			fmt.Sprintf("%.6f", result.Avg),
			fmt.Sprintf("%.6f", result.StdDev),
			fmt.Sprintf("%.6f", result.U_A),
			fmt.Sprintf("%.6f", result.U_B),
			fmt.Sprintf("%.6f", result.U),
			fmt.Sprintf("%.6f ± %.6f", result.Avg, result.U),
		}

		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// 显示错误页面
func showErrorPage(w http.ResponseWriter, errorMessage string) {
	tmpl := template.Must(template.New("error").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>错误 - 物理实验不确定度计算</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 20px;
				background-color: #f5f5f5;
			}
			.container {
				max-width: 800px;
				margin: 0 auto;
				background-color: white;
				padding: 30px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			}
			h1 {
				color: #f44336;
				text-align: center;
				margin-bottom: 30px;
			}
			.error-message {
				background-color: #ffebee;
				padding: 20px;
				border-left: 4px solid #f44336;
				margin-bottom: 30px;
				font-size: 16px;
			}
			.back-link {
				display: inline-block;
				padding: 12px 24px;
				background-color: #4CAF50;
				color: white;
				text-decoration: none;
				border-radius: 4px;
				font-weight: bold;
				font-size: 16px;
			}
			.back-link:hover {
				background-color: #45a049;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>输入错误</h1>
			<div class="error-message">
				{{.ErrorMessage}}
			</div>
			<a href="/" class="back-link">返回首页</a>
		</div>
	</body>
	</html>
	`))

	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMessage,
	}

	tmpl.Execute(w, data)
}

// 处理计算请求
func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 解析表单数据
	r.ParseForm()

	// 获取参数
	groupCountStr := r.FormValue("groupCount")
	measureCountStr := r.FormValue("measureCount")
	deltaStr := r.FormValue("delta")

	// 转换参数
	groupCount, err := strconv.Atoi(groupCountStr)
	if err != nil || groupCount <= 0 {
		showErrorPage(w, "无效的组数，请输入正整数")
		return
	}

	measureCount, err := strconv.Atoi(measureCountStr)
	if err != nil || measureCount <= 0 {
		showErrorPage(w, "无效的测量次数，请输入正整数")
		return
	}

	delta, err := strconv.ParseFloat(deltaStr, 64)
	if err != nil {
		showErrorPage(w, "无效的误差限，请输入有效的数字")
		return
	}

	// 处理每组数据
	results := make([]Result, 0, groupCount)

	for group := 1; group <= groupCount; group++ {
		dataStr := r.FormValue(fmt.Sprintf("data_%d", group))
		parts := strings.Fields(dataStr)

		if len(parts) < measureCount {
			showErrorPage(w, fmt.Sprintf("第%d组数据不足，需要输入%d个测量值", group, measureCount))
			return
		}

		measurements := make([]float64, measureCount)
		for i := 0; i < measureCount; i++ {
			val, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				showErrorPage(w, fmt.Sprintf("第%d组第%d个值无效，请输入有效的数字", group, i+1))
				return
			}
			measurements[i] = val
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
	}

	// 导出为CSV
	filePath, err := exportToCSV(results, delta)
	if err != nil {
		showErrorPage(w, "导出失败，请重试")
		return
	}

	// 渲染结果页面
	tmpl := template.Must(template.New("result").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>物理实验不确定度计算结果</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 20px;
				background-color: #f5f5f5;
			}
			.container {
				max-width: 1000px;
				margin: 0 auto;
				background-color: white;
				padding: 20px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			}
			h1 {
				color: #333;
				text-align: center;
			}
			h2 {
				color: #555;
				margin-top: 30px;
			}
			.table-container {
				overflow-x: auto;
				margin: 20px 0;
			}
			table {
				width: 100%;
				border-collapse: collapse;
			}
			th, td {
				padding: 10px;
				text-align: left;
				border-bottom: 1px solid #ddd;
			}
			th {
				background-color: #f2f2f2;
				font-weight: bold;
			}
			tr:hover {
				background-color: #f5f5f5;
			}
			.data-cell {
				white-space: pre-wrap;
			}
			.export-link {
				display: inline-block;
				margin-top: 20px;
				padding: 10px 20px;
				background-color: #4CAF50;
				color: white;
				text-decoration: none;
				border-radius: 4px;
				font-weight: bold;
			}
			.export-link:hover {
				background-color: #45a049;
			}
			.back-link {
				display: inline-block;
				margin-top: 20px;
				margin-left: 20px;
				padding: 10px 20px;
				background-color: #f44336;
				color: white;
				text-decoration: none;
				border-radius: 4px;
				font-weight: bold;
			}
			.back-link:hover {
				background-color: #da190b;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>物理实验不确定度计算结果</h1>
			
			{{range .Results}}
			<h2>第 {{.GroupNum}} 组数据</h2>
			<div class="table-container">
				<table>
					<tr>
						<th>测量数据</th>
						<td class="data-cell">{{range .Data}}{{printf "%.4f " .}}{{end}}</td>
					</tr>
					<tr>
						<th>平均值 L</th>
						<td>{{printf "%.6f" .Avg}}</td>
					</tr>
					<tr>
						<th>标准偏差 s</th>
						<td>{{printf "%.6f" .StdDev}}</td>
					</tr>
					<tr>
						<th>A类不确定度 U_A</th>
						<td>{{printf "%.6f" .U_A}}</td>
					</tr>
					<tr>
						<th>B类不确定度 U_B</th>
						<td>{{printf "%.6f" .U_B}}</td>
					</tr>
					<tr>
						<th>合成不确定度 U</th>
						<td>{{printf "%.6f" .U}}</td>
					</tr>
					<tr>
						<th>最终测量结果</th>
						<td>{{printf "%.6f ± %.6f" .Avg .U}}</td>
					</tr>
				</table>
			</div>
			{{end}}
			
			<h2>导出结果</h2>
			<p>结果已导出到：{{.FilePath}}</p>
			<a href="/" class="back-link">返回首页</a>
		</div>
	</body>
	</html>
	`))

	data := struct {
		Results  []Result
		FilePath string
	}{
		Results:  results,
		FilePath: filePath,
	}

	tmpl.Execute(w, data)
}

// 处理首页请求
func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>物理实验不确定度计算</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 20px;
				background-color: #f5f5f5;
			}
			.container {
				max-width: 800px;
				margin: 0 auto;
				background-color: white;
				padding: 30px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			}
			h1 {
				color: #333;
				text-align: center;
				margin-bottom: 30px;
			}
			.form-group {
				margin-bottom: 20px;
			}
			label {
				display: block;
				margin-bottom: 8px;
				font-weight: bold;
				color: #555;
			}
			input[type="text"] {
				width: 100%;
				padding: 10px;
				border: 1px solid #ddd;
				border-radius: 4px;
				font-size: 16px;
			}
			input[type="submit"] {
				background-color: #4CAF50;
				color: white;
				padding: 12px 20px;
				border: none;
				border-radius: 4px;
				cursor: pointer;
				font-size: 16px;
				font-weight: bold;
				width: 100%;
				margin-top: 20px;
			}
			input[type="submit"]:hover {
				background-color: #45a049;
			}
			.data-group {
				margin-top: 30px;
				padding: 20px;
				border: 1px solid #ddd;
				border-radius: 4px;
				background-color: #f9f9f9;
			}
			.data-group h3 {
				margin-top: 0;
				color: #333;
			}
			.example {
				font-style: italic;
				color: #666;
				margin-top: 5px;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>物理实验不确定度计算</h1>
			<form method="post" action="/calculate">
				<div class="form-group">
					<label for="groupCount">实验数据的组数</label>
					<input type="text" id="groupCount" name="groupCount" placeholder="例如：3 表示有3组数据" required>
					<div class="example">示例：3</div>
				</div>
				
				<div class="form-group">
					<label for="measureCount">每组的测量次数</label>
					<input type="text" id="measureCount" name="measureCount" placeholder="例如：6 表示每组6次测量" required>
					<div class="example">示例：6</div>
				</div>
				
				<div class="form-group">
					<label for="delta">仪器的误差限 (delta)</label>
					<input type="text" id="delta" name="delta" placeholder="例如：0.01" required>
					<div class="example">示例：0.01</div>
				</div>
				
				<script>
					// 动态生成数据输入字段
					document.getElementById('groupCount').addEventListener('input', function() {
						const groupCount = parseInt(this.value) || 0;
						const dataContainer = document.getElementById('dataContainer');
						dataContainer.innerHTML = '';
						
						for (let i = 1; i <= groupCount; i++) {
							const dataGroup = document.createElement('div');
							dataGroup.className = 'data-group';
							dataGroup.innerHTML = 
								'<h3>第 ' + i + ' 组数据</h3>' +
								'<label for="data_' + i + '">' + (document.getElementById('measureCount').value || 6) + ' 次测量的数据（用空格分隔）</label>' +
								'<input type="text" id="data_' + i + '" name="data_' + i + '" placeholder="例如：1.23 1.25 1.24 1.26 1.23 1.25" required>' +
								'<div class="example">示例：1.23 1.25 1.24 1.26 1.23 1.25</div>';
							dataContainer.appendChild(dataGroup);
						}
					});
					
					document.getElementById('measureCount').addEventListener('input', function() {
						const measureCount = this.value;
						const dataGroups = document.querySelectorAll('.data-group');
						dataGroups.forEach((group, index) => {
							const label = group.querySelector('label');
							label.textContent = measureCount + ' 次测量的数据（用空格分隔）';
						});
					});
				</script>
				
				<div id="dataContainer"></div>
				
				<input type="submit" value="计算并导出结果">
			</form>
		</div>
	</body>
	</html>
	`))

	tmpl.Execute(w, nil)
}

func main() {
	// 注册路由
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/calculate", handleCalculate)

	// 启动服务器
	port := "8082"
	fmt.Printf("服务器已启动，访问地址：http://localhost:%s\n", port)
	fmt.Println("程序已启动，打开浏览器访问上面的地址开始使用")
	fmt.Println("按 Ctrl+C 停止服务器")

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("服务器启动失败：%v\n", err)
	}
}
