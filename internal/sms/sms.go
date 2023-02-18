package sms

import (
	"finalwork/internal/countries"
	"fmt"
	"os"
	"strings"
	//"io/ioutil"
)

type SMSData struct {
	Country      string `json:"country"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
	Provider     string `json:"provider"`
}

var providers = map[string]struct{}{ // создадим мап с ключами, которые соответствуют названиям допустимых провайдеров
	"Topolo": {},
	"Rond":   {},
	"Kildy":  {},
}

type SmsCountryRepository countries.CountryRepository // локальная обёртка над типом countries.CountryRepository

func openFile(fileName string) ([]byte, error) { // функция открытия файла
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fi, _ := file.Stat()
	buf := make([]byte, fi.Size())
	if _, err = file.Read(buf); err != nil { // читаем, получаем слайс байтов (используем вместо ioutil)
		fmt.Println("File reading error:", err)
		panic(err)
	}
	return buf, nil
}

func (r *SmsCountryRepository) GetSmsData(fileName string) ([]SMSData, error) { // функция сбора данных о системе SMS
	var SMSDataSlice []SMSData
	fileByteSlice, err := openFile(fileName) // получаем слайс байтов из файла, через отдельную функцию
	if err != nil {
		return SMSDataSlice, err
	}
	sep := "\n"                                                  // создаём сепаратор
	fileStringSlice := strings.Split(string(fileByteSlice), sep) // разделяем весь текст на слайс подстрок по "\n"
	for _, v := range fileStringSlice {                          // проходим по каждой подстроке
		sep := ";"                                 // создаём сепаратор
		singleStringSlice := strings.Split(v, sep) // разделяем подстроку по разделителю ";"
		if len(singleStringSlice) < 4 {            // проверяем, что кол-во элементов не меньше 4
			continue
		} else {
			SMSDataStruct, ok := r.parseStringSlice(singleStringSlice) // парсинг в структуру и проверка требованиям
			if ok {                                                    //если true
				SMSDataSlice = append(SMSDataSlice, SMSDataStruct) // добавление структуры в результирующий срез
			}
		}
	}
	return SMSDataSlice, nil
}

// функция создания структуры из слайса строк и проверки поля Country по коду alpha-2
func (r *SmsCountryRepository) parseStringSlice(singleStringSlice []string) (SMSData, bool) {
	SMSds := SMSData{
		Country:      singleStringSlice[0],
		Bandwidth:    singleStringSlice[1],
		ResponseTime: singleStringSlice[2],
		Provider:     singleStringSlice[3],
	}
	if _, ok := r.CountryByCode[countries.Code(SMSds.Country)]; ok { // проверяем по alpha-2, обращаясь к хранилищу SmsCountryRepository по ключу = значению поля Country у полученной структуры
		if _, ok := providers[SMSds.Provider]; ok { // проверяем провайдера, обращаясь к мап по ключу = значению поля Provider у полученной структуры
			return SMSds, true // если значения по ключам в мап имеются, возвращаем структуру и true
		}
	}
	return SMSds, false // возвращаем структуру и false

	// if !ok {
	//      return SMSds, false
	// }
	// if _, ok := providers[SMSds.Provider]; !ok {
	//      return SMSds, false
	// }
	// return SMSds, true
}
