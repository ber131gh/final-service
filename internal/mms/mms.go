package mms

import (
	"encoding/json"
	"finalwork/internal/countries"
	"io"

	"net/http"
)

type MMSData struct {
	Country      string `json:"country"`
	Provider     string `json:"provider"`
	Bandwidth    string `json:"bandwidth"`
	ResponseTime string `json:"response_time"`
}

var providers = map[string]struct{}{ // создадим мап с ключами, которые соответствуют названиям допустимых провайдеров
	"Topolo": {},
	"Rond":   {},
	"Kildy":  {},
}

type MmsCountryRepository countries.CountryRepository // локальная обёртка над типом countries.CountryRepository

func (r *MmsCountryRepository) GetMmsData(addr string) ([]MMSData, int, error) { // функция сбора данных о системе MMS
	var MMSDataSlice []MMSData
	client := &http.Client{}      // создаем структуру Client
	resp, err := client.Get(addr) // отправляем GET-запрос по addr
	if err != nil {
		return MMSDataSlice, resp.StatusCode, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 { // проверяем код ответа API
		body, err := io.ReadAll(resp.Body) // считываем тело запроса
		if err != nil {
			return MMSDataSlice, resp.StatusCode, err
		}
		if err := json.Unmarshal(body, &MMSDataSlice); err != nil { // используем функцию Unmarshal
			return MMSDataSlice, resp.StatusCode, err
		}
		MMSDataSlice = r.checkSliceByOptions(MMSDataSlice) // проверка слайса []MMSData требованиям
		return MMSDataSlice, resp.StatusCode, nil
	}
	return MMSDataSlice, resp.StatusCode, nil
}

func (r *MmsCountryRepository) checkSliceByOptions(MMSDataSlice []MMSData) []MMSData {
	for i, v := range MMSDataSlice { // проходим по слайсу []MMSData
		if _, ok := r.CountryByCode[countries.Code(v.Country)]; ok { // проверяем по alpha-2, обращаясь к хранилищу MmsCountryRepository по ключу = значению поля Country каждой отдельной структуры слайса []MMSData
			if _, ok := providers[v.Provider]; ok { // проверяем провайдера, обращаясь к мап по ключу = значению поля Provider каждой отдельной структуры слайса []MMSData
				continue // если значение по ключу в мап имеется, то сбрасываем данную итерацию цикла
			}
		} // если значения по ключу в мап не имеется, то удаляем данную структуру из слайса []MMSData
		MMSDataSlice[len(MMSDataSlice)-1], MMSDataSlice[i] = MMSDataSlice[i], MMSDataSlice[len(MMSDataSlice)-1]
		MMSDataSlice = MMSDataSlice[:len(MMSDataSlice)-1]
	}
	return MMSDataSlice // возвращаем слайс структур
}
