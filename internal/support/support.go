package support

import (
	"encoding/json"
	"io"
	"net/http"
)

type SupportData struct {
	Topic         string `json:"topic"`
	ActiveTickets int    `json:"active_tickets"`
}

func GetSupportData(addr string) ([]SupportData, int, error) { // функция сбора данных о системе support
	var supportDataSlice []SupportData
	resp, err := http.Get(addr) //отправляем GET-запрос по addr
	if err != nil {
		return supportDataSlice, resp.StatusCode, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 { // проверяем код ответа API
		body, err := io.ReadAll(resp.Body) // считываем тело запроса
		if err != nil {
			return supportDataSlice, resp.StatusCode, err
		}
		if err := json.Unmarshal(body, &supportDataSlice); err != nil { // используем функцию Unmarshal
			return supportDataSlice, resp.StatusCode, err
		}
		return supportDataSlice, resp.StatusCode, nil
	}
	return supportDataSlice, resp.StatusCode, nil
}
