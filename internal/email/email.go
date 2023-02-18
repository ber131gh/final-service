package email

import (
	"finalwork/internal/countries"
	"io"
	"strconv"
	"strings"

	"log"
	"os"
)

type EmailData struct {
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	DeliveryTime int    `json:"delivery_time"`
}

type EmailCountryRepository countries.CountryRepository // локальная обёртка над типом countries.CountryRepository

var validProviders = map[string]struct{}{ // создадим мап с ключами, которые соответствуют названиям допустимых провайдеров
	"Gmail":       {},
	"Yahoo":       {},
	"Hotmail":     {},
	"MSN":         {},
	"Orange":      {},
	"Comcast":     {},
	"AOL":         {},
	"Live":        {},
	"RediffMail":  {},
	"GMX":         {},
	"Proton Mail": {},
	"Yandex":      {},
	"Mail.ru":     {},
}

func (r *EmailCountryRepository) GetEmailData(fileName string) ([]EmailData, error) { // функция сбора данных о системе Email
	var emailDataSlice []EmailData
	file, err := os.Open(fileName) // открываем файл
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file) // читаем, получаем слайс байтов (используем вместо ioutil)
	if err != nil {
		log.Panic(err)
	}
	sep := "\n"                                      // создаём сепаратор
	stringSplit := strings.Split(string(bytes), sep) // разделяем весь текст на слайс подстрок по "\n"
	for _, v := range stringSplit {                  // проходим по каждой подстроке
		sep := ";"                                 // создаём сепаратор
		singleStringSlice := strings.Split(v, sep) // разделяем подстроку по разделителю ";"
		if len(singleStringSlice) < 3 {            // проверяем, что кол-во элементов не меньше 3
			continue
		} else {
			emailDataStruct, ok := r.parseStringleSlice(singleStringSlice) // парсинг в структуру и проверка требованиям
			if ok {                                                        //если true
				emailDataSlice = append(emailDataSlice, emailDataStruct) // добавление структуры в результирующий срез
			}
		}
	}
	return emailDataSlice, nil
}

// функция создания структуры из слайса строк и проверки поля Country по коду alpha-2
func (r *EmailCountryRepository) parseStringleSlice(s []string) (EmailData, bool) {
	var emailDataStruct EmailData
	emailDataStruct.Country = s[0]
	emailDataStruct.Provider = s[1]
	eDt, err := strconv.Atoi(s[2]) // конвертируем строку в int
	if err != nil {
		return emailDataStruct, false
	}
	emailDataStruct.DeliveryTime = eDt // присваиваем значение числовому полю структуры EmailData

	if _, ok := r.CountryByCode[countries.Code(emailDataStruct.Country)]; ok { // проверяем по alpha-2, обращаясь к хранилищу emailCountryRepository по ключу = значению поля Country у полученной структуры
		if _, ok := validProviders[emailDataStruct.Provider]; ok { // проверяем провайдера, обращаясь к мап по ключу = значению поля Provider у полученной структуры
			return emailDataStruct, true // если значения по ключам в мап имеются, возвращаем структуру и true
		}
	}
	return emailDataStruct, false // возвращаем структуру и false
}
