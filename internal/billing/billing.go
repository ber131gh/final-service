package billing

import (
	"io"
	"os"
	"strconv"
)

type BillingData struct {
	CreateCustomer bool `json:"create_customer"`
	Purchase       bool `json:"purchase"`
	Payout         bool `json:"payout"`
	Recurring      bool `json:"recurring"`
	FraudControl   bool `json:"fraud_control"`
	CheckoutPage   bool `json:"checkout_page"`
}

func GetBillingData(fileName string) (BillingData, error) { // функция сбора данных о системе Billing
	var billingDataStruct BillingData
	file, err := os.Open(fileName) // открываем файл
	if err != nil {
		return billingDataStruct, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file) // читаем, получаем слайс байтов (используем вместо ioutil)
	if err != nil {
		return billingDataStruct, err
	}
	//fmt.Println(string(bytes)) // печатает маску в виде строки

	number, err := strconv.ParseInt(string(bytes), 2, 8) // интерпретируем строку в число
	if err != nil {
		return billingDataStruct, err
	}
	mNumber := uint8(number) // присваиваем значение переменной с типом uint8.
	//fmt.Println(mNumber) // печатает проверочное число
	var boolSlice [6]bool // создаем массив длиной 6, равной количеству полей структуры BillingData
	for i := 0; i < 6; i++ {
		if mNumber&(1<<i) != 0 { // используем оператор битовых вычислений. Если в обоих числах этот бит есть, то он будет установлен
			boolSlice[i] = true
		} else {
			boolSlice[i] = false
		}
	}
	billingDataStruct = BillingData{ // заполним поля структуры BillingData значениями элементов массива boolSlice
		boolSlice[0],
		boolSlice[1],
		boolSlice[2],
		boolSlice[3],
		boolSlice[4],
		boolSlice[5],
	}
	return billingDataStruct, nil
}
