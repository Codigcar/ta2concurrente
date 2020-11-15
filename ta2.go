package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func presicion(data_de_entrada [16][2]float64, pesos [2]float64, sesgo float64, data_de_salida [16]int, presicion_ch chan float64) {
	numero_correcto := 0
	numero_incorrecto := 0
	salida_ch := make(chan int)

	for i := 0; i < 16; i++ {
		go calcularSalida(data_de_entrada[i], pesos, sesgo, salida_ch)
		salida := <-salida_ch
		if salida == data_de_salida[i] {
			numero_correcto++
		} else {
			numero_incorrecto++
		}
	}
	resultado := (float64(numero_correcto) * 1.0) / (float64(numero_correcto) + float64(numero_incorrecto))
	presicion_ch <- resultado
}
func error(data_de_entrada [2]float64, pesos [2]float64, sesgo float64, data_de_salida int, error_ch chan float64) {
	suma := 0.0
	Y2 := float64(data_de_salida)
	for j := 0; j < len(data_de_entrada); j++ {
		suma += data_de_entrada[j] * pesos[j]
	}
	suma += sesgo
	_error := 0.5 * (suma - Y2) * (suma - Y2)

	error_ch <- _error
}
func errorTotal(data_de_entrada [16][2]float64, pesos [2]float64, sesgo float64, data_de_salida [16]int, error_total_ch chan float64) {
	error_total := 0.0
	error_ch := make(chan float64)

	for i := 0; i < len(data_de_entrada); i++ {
		go error(data_de_entrada[i], pesos, sesgo, data_de_salida[i], error_ch)
		error_total += <-error_ch

	}
	error_total_ch <- error_total
}
func esPositivo(x float64, salida_ch chan int) {
	if x >= 0.0 {
		salida_ch <- +1
	} else {
		salida_ch <- -1
	}
}
func calcularSalida(data_de_entrada [2]float64, pesos [2]float64, sesgo float64, salida_ch chan int) {
	resultado := 0.0
	for j := 0; j < len(data_de_entrada); j++ {
		resultado += data_de_entrada[j] * pesos[j]
	}
	resultado += sesgo
	go esPositivo(resultado, salida_ch)
}

func entrenamiento(data_de_entrada [16][2]float64, tasa_aprendizaje float64, maxIteraciones int, data_de_salida [16]int, pesos_ch chan [2]float64, sesgo_ch chan float64) {
	var pesos_mejor [2]float64
	var pesos [2]float64
	sesgo_mejor := 0.0
	error_mejor := 100.0
	iteracion := 0
	sesgo := 0.0
	salida_ch := make(chan int)
	error_total_ch := make(chan float64)

	for iteracion < maxIteraciones {
		for i := 0; i < 16; i++ {
			go calcularSalida(data_de_entrada[i], pesos, sesgo, salida_ch)
			salida_real := <-salida_ch
			salida_deseada := data_de_salida[i]

			if salida_real != salida_deseada {
				delta := float64(salida_deseada - salida_real)
				for j := 0; j < 2; j++ {
					pesos[j] = pesos[j] + (tasa_aprendizaje * delta * data_de_entrada[i][j])
				}
				sesgo = sesgo + (tasa_aprendizaje * delta)
				go errorTotal(data_de_entrada, pesos, sesgo, data_de_salida, error_total_ch)
				totalError := <-error_total_ch

				if totalError < error_mejor {
					error_mejor = totalError
					pesos_mejor = pesos
					sesgo_mejor = sesgo

					//fmr.Println("%error: ", error_mejor)
					//fmr.Println("sesgo: ", sesgo_mejor)
					//fmr.Println("pesos: ", pesos_mejor)

				}
			}
		}

		iteracion++
	}
	pesos_ch <- pesos_mejor
	sesgo_ch <- sesgo_mejor
	return
}

func string_a_Float(data_ch chan []float64, columna []string) {
	var data []float64
	for _, elem := range columna {
		i, err := strconv.ParseFloat(elem, 64)
		if err == nil {
			data = append(data, i)
		}
	}
	data_ch <- data
}

func string_a_Int(data_ch chan []int, columna []string) {
	var data []int
	for _, elem := range columna {
		i, err := strconv.Atoi(elem)
		if err == nil {
			data = append(data, i)
		}
	}
	data_ch <- data
}
func leerData(data_x1_ch chan []float64, data_x2_ch chan []float64, data_y_ch chan []int) {

	var columna_x1 []string
	var columna_x2 []string
	var columna_y []string

	resp, _ := http.Get("https://raw.githubusercontent.com/Codigcar/TA2Concu/master/ta2/data4.csv?token=AIGLPDAMNEI4GQMWQ55JHK27W52AQ")
	data, _ := ioutil.ReadAll(resp.Body)
	s := strings.Split(string(data), ",")

	for i := 1; i < len(s); i += 4 {
		columna_x1 = append(columna_x1, s[i])
		columna_x2 = append(columna_x2, s[i+1])
		columna_y = append(columna_y, s[i+2])
	}

	go string_a_Float(data_x1_ch, columna_x1)
	go string_a_Float(data_x2_ch, columna_x2)
	go string_a_Int(data_y_ch, columna_y)

}
func mostrarDataDeEntrenamiento(data_x1 []float64, data_x2 []float64, data_y []int) {
	for i := 0; i < len(data_x1); i++ {
		fmt.Println(data_x1[i], " , ", data_x2[i], " -> ", data_y[i])
	}
}
func main() {

	// data para el entrenamiento
	var data_de_entrada [16][2]float64
	var data_de_salida [16]int
	tasa_aprendizaje := 0.001
	maxIteraciones := 500
	data_x1_ch := make(chan []float64)
	data_x2_ch := make(chan []float64)
	data_y_ch := make(chan []int)
	pesos_ch := make(chan [2]float64)
	sesgo_ch := make(chan float64)
	presicion_ch := make(chan float64)
	salida_ch := make(chan int)

	go leerData(data_x1_ch, data_x2_ch, data_y_ch)
	data_x1, data_x2, data_y := <-data_x1_ch, <-data_x2_ch, <-data_y_ch
	for i := 0; i < 16; i++ {
		data_de_entrada[i] = [2]float64{data_x1[i], data_x2[i]}
		data_de_salida[i] = data_y[i]
	}
	go entrenamiento(data_de_entrada, tasa_aprendizaje, maxIteraciones, data_de_salida, pesos_ch, sesgo_ch)
	pesos, sesgo := <-pesos_ch, <-sesgo_ch
	go presicion(data_de_entrada, pesos, sesgo, data_de_salida, presicion_ch)
	presicion := <-presicion_ch
	mostrarDataDeEntrenamiento(data_x1, data_x2, data_y)
	fmt.Println("Mejor peso: ", pesos)
	fmt.Println("Mejor sesgo: ", sesgo)
	fmt.Println("Mejor presicion: ", presicion*100, "%")

	fmt.Println("Red Neuronal Entrenado Completo")
	fmt.Println("---------------------")
	var data_desconocida_test = [2]float64{-1.0, 10.5}
	fmt.Println("Nuevo valor a ingresar: ", data_desconocida_test)
	go calcularSalida(data_desconocida_test, pesos, sesgo, salida_ch)
	resultado_predicho := <-salida_ch
	fmt.Println("Resultado por perceptron: ", resultado_predicho)

}
