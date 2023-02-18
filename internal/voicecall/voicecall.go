package voicecall

import (
	"finalwork/internal/countries"
	"io"
	"strconv"
	"strings"

	"log"
	"os"
)

type VoiceData struct {
	Country             string  `json:"country"`
	CurrentLoad         int     `json:"current_load"`
	ResponseTime        int     `json:"response_time"`
	Provider            string  `json:"provider"`
	ConnectionStability float32 `json:"connection_stability"`
	PurityTTFB          int     `json:"purity_ttfb"`
	CallDuration        int     `json:"call_duration"`
	UnknownField        int     `json:"unknown_field"` // обозначил как "неизвестное поле", т.к в описании ("Этап 4 п.5")  в перечне 7 полей, а не 8. Из файла Voicedata видно, что оно является числовым типом.
}

type VoiceCountryRepository countries.CountryRepository // локальная обёртка над типом countries.CountryRepository

var validProviders = map[string]struct{}{ // создадим мап с ключами, которые соответствуют названиям допустимых провайдеров
	"TransparentCalls": {},
	"E-Voice":          {},
	"JustPhone":        {},
}

func (r *VoiceCountryRepository) GetVoiceData(fileName string) ([]VoiceData, error) { // функция сбора данных о системе Voicecall
	var voiceDataSlice []VoiceData
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
		if len(singleStringSlice) < 8 {            // проверяем, что кол-во элементов не меньше 8
			continue
		} else {
			voiceDataStruct, ok := r.parseStringleSlice(singleStringSlice) // парсинг в структуру и проверка требованиям
			if ok {                                                        //если true
				voiceDataSlice = append(voiceDataSlice, voiceDataStruct) // добавление структуры в результирующий срез
			}
		}
	}
	return voiceDataSlice, nil
}

// функция создания структуры из слайса строк и проверки поля Country по коду alpha-2
func (r *VoiceCountryRepository) parseStringleSlice(s []string) (VoiceData, bool) {
	fAtoi := func(s string) (int, error) { // функция конвертации строки в int
		intNumb, err := strconv.Atoi(s)
		if err != nil {
			return intNumb, err
		}
		return intNumb, nil
	}
	fAtof := func(s string) (float32, error) { // функция конвертации строки в float32
		intNumb, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return float32(intNumb), err
		}
		return float32(intNumb), nil
	}

	var voiceDataStruct VoiceData // создаем структуру типа VoiceData
	var err error
	voiceDataStruct.Country = s[0]
	voiceDataStruct.CurrentLoad, err = fAtoi(s[1])
	if err != nil { // если строка повреждена, возвращаем false. Такие строки пропускаются и не добавляюся в слайс []Voicecall
		return voiceDataStruct, false
	}
	voiceDataStruct.ResponseTime, err = fAtoi(s[2])
	if err != nil {
		return voiceDataStruct, false
	}
	voiceDataStruct.Provider = s[3]
	voiceDataStruct.ConnectionStability, err = fAtof(s[4])
	if err != nil {
		return voiceDataStruct, false
	}
	voiceDataStruct.PurityTTFB, err = fAtoi(s[5])
	if err != nil {
		return voiceDataStruct, false
	}
	voiceDataStruct.CallDuration, err = fAtoi(s[6])
	if err != nil {
		return voiceDataStruct, false
	}
	voiceDataStruct.UnknownField, err = fAtoi(s[7])
	if err != nil {
		return voiceDataStruct, false
	}

	if _, ok := r.CountryByCode[countries.Code(voiceDataStruct.Country)]; ok { // проверяем по alpha-2, обращаясь к хранилищу VoiceCountryRepository по ключу = значению поля Country у полученной структуры
		if _, ok := validProviders[voiceDataStruct.Provider]; ok { // проверяем провайдера, обращаясь к мап по ключу = значению поля Provider у полученной структуры
			return voiceDataStruct, true // если значения по ключам в мап имеются, возвращаем структуру и true
		}
	}
	return voiceDataStruct, false // возвращаем структуру и false
}
