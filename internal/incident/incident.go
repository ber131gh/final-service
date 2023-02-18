package incident

import (
	"encoding/json"
	"io"
	"net/http"
)

type IncidentData struct {
	Topic  string `json:"topic"`
	Status string `json:"status"`
}

func GetIncidentData(addr string) ([]IncidentData, int, error) { // функция сбора данных о системе Incident
	var incidentDataSlice []IncidentData
	resp, err := http.Get(addr) //отправляем GET-запрос по addr
	if err != nil {
		return incidentDataSlice, resp.StatusCode, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 { // проверяем код ответа API
		body, err := io.ReadAll(resp.Body) // считываем тело запроса
		if err != nil {
			return incidentDataSlice, resp.StatusCode, err
		}
		if err := json.Unmarshal(body, &incidentDataSlice); err != nil { // используем функцию Unmarshal
			return incidentDataSlice, resp.StatusCode, err
		}
		return incidentDataSlice, resp.StatusCode, nil
	}
	return incidentDataSlice, resp.StatusCode, nil
}
