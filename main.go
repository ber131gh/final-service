package main

import (
	"context"
	"encoding/json"
	"finalwork/internal/billing"
	"finalwork/internal/countries"
	"finalwork/internal/email"
	"finalwork/internal/incident"
	"finalwork/internal/mms"
	"finalwork/internal/sms"
	"finalwork/internal/support"
	"finalwork/internal/voicecall"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

type ( // локальные типы над типами данных из других пакетов
	SMSData       sms.SMSData
	MMSData       mms.MMSData
	VoiceCallData voicecall.VoiceData
	EmailData     email.EmailData
	BillingData   billing.BillingData
	SupportData   support.SupportData
	IncidentData  incident.IncidentData
)

type ResultT struct { // конечная родительская структура
	Status bool       `json:"status"` // True, если все этапы сбора данных прошли успешно, False во всех остальных случаях
	Data   ResultSetT `json:"data"`   // Заполнен, если все этапы сбора  данных прошли успешно, nil во всех остальных случаях
	Error  string     `json:"error"`  // Пустая строка, если все этапы сбора данных прошли успешно, в случае ошибки заполнено текстом ошибки
}

type ResultSetT struct { // родительская структура (отфильтрованный результат сбора данных по различным системам)
	SMS       [][]SMSData              `json:"sms"`
	MMS       [][]MMSData              `json:"mms"`
	VoiceCall []VoiceCallData          `json:"voice_call"`
	Email     map[string][][]EmailData `json:"email"`
	Billing   BillingData              `json:"billing"`
	Support   []int                    `json:"support"`
	Incidents []IncidentData           `json:"incident"`
}

// файлы содержащие адреса получения данных из файлов и через API
const smsFileName = "simulator/skillbox-diploma/sms.data"
const mmsUrlAddr string = "http://127.0.0.1:8383/mms"
const voiceFileName = "simulator/skillbox-diploma/voice.data"
const emailFileName = "simulator/skillbox-diploma/email.data"
const billingFileName = "simulator/skillbox-diploma/billing.data"
const supportUrlAddr string = "http://127.0.0.1:8383/support"
const incidentUrlAddr string = "http://127.0.0.1:8383/accendent"

var (
	countryRepo      = countries.ISOCountryRepository()              // создаем хранилище типа CountryRepository(мапа с ключом==alpha2 и значением==названию страны)
	smsCountryRepo   = sms.SmsCountryRepository(countryRepo)         // обертка над countryRepo
	mmsCountryRepo   = mms.MmsCountryRepository(countryRepo)         // обертка над countryRepo
	voiceCountryRepo = voicecall.VoiceCountryRepository(countryRepo) // обертка над countryRepo
	emailCountryRepo = email.EmailCountryRepository(countryRepo)     // обертка над countryRepo
)

func main() {
	r := mux.NewRouter()                           // создаем роутер
	r.HandleFunc("/", handleConnection)            // добавляем к роутеру обработку функции handleConnection
	r.HandleFunc("/systemsstatus", getSystemsData) // добавляем к роутеру обработку функции getSystemsData
	server := http.Server{                         // создаем сервер
		Addr:    "localhost:8282", // адрес для прослушивания
		Handler: r,                // роутер
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() { // горутина закрывающая
		for {
			s := <-sigChan // ожидаем сигнала os.Interrupt
			fmt.Println("Сигнал:", s)
			fmt.Println("Выходим из программы")
			if err := server.Shutdown(context.Background()); err != nil { // закрываем сервер
				fmt.Printf("Server shutdown error: %s\n", err)
			}
		}
	}()
	fmt.Println("Starting server on localhost: 8282")
	if err := server.ListenAndServe(); err != nil { // запускаем сервер
		fmt.Println("ListenAndServe:", err)
	}
}

func handleConnection(w http.ResponseWriter, r *http.Request) { // функция первичного тестирования обработки запросов
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK") // возвращаем в ответ "OK"
}

func getSystemsData(w http.ResponseWriter, r *http.Request) { // функция возвращающая в Response, конечную структуру с отфильтрованными данными в формате json.
	if r.Method == "GET" {
		systemData := getResultT() // вызываем функцию получения конечной родительской структуры
		//fmt.Fprintf(w, "status: %v", systemData)
		csD, err := json.Marshal(systemData) // конвертация структуры в json ([]byte)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(csD) // записываем данные в Response
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (r *ResultSetT) getAndSortSMS() error { // функция фильтрации данных системы SMS
	smsData, err := smsCountryRepo.GetSmsData(smsFileName) // получаем данные из системы SMS
	if err != nil {
		return err
	} else {
		var smsDSetCountry []SMSData // создаем слайс типа SMSData
		for _, v := range smsData {  // проходим по слайсу данных системы SMS
			v.Country = smsCountryRepo.CountryByCode[countries.Code(v.Country)].Name // в каждом элементе заменяем значение поля кода страны на полное название страны из хранилища SmsCountryRepository
			smsDSetCountry = append(smsDSetCountry, SMSData(v))                      // добавляем в слайс smsDSetCountry каждый обновленный элемент слайса системы SMS, приводя его к типу SMSData
		}
		var providerSortedSlice = make([]SMSData, len(smsData)) // инициализируем слайс, для дальнейшей сортировки по имени провайдера
		copy(providerSortedSlice, smsDSetCountry)               // копируем в него элементы из слайса типа SMSData, с полными именами городов
		for i := 0; i < len(providerSortedSlice); i++ {         // для сортировки используем "сортировку выбора"
			var minIdx = i
			for j := i; j < len(providerSortedSlice); j++ { // для каждого i, ищем минимальный элемент в правой части
				if providerSortedSlice[j].Provider < providerSortedSlice[minIdx].Provider {
					minIdx = j
				}
			}
			providerSortedSlice[i], providerSortedSlice[minIdx] = providerSortedSlice[minIdx], providerSortedSlice[i] // переставляем мин.элемент налево
		}
		var countrySortedSlice = make([]SMSData, len(smsData)) // инициализируем слайс, для дальнейшей сортировки по имени страны
		copy(countrySortedSlice, smsDSetCountry)
		for i := 0; i < len(countrySortedSlice); i++ {
			var minIdx = i
			for j := i; j < len(countrySortedSlice); j++ {
				if countrySortedSlice[j].Country < countrySortedSlice[minIdx].Country {
					minIdx = j
				}
			}
			countrySortedSlice[i], countrySortedSlice[minIdx] = countrySortedSlice[minIdx], countrySortedSlice[i]
		}
		r.SMS = [][]SMSData{ // заполняем поле SMS структуры ResultSetT. Инициализируем двумерный слайс и присваиваем значения
			providerSortedSlice,
			countrySortedSlice,
		}
		// fmt.Println("SMS system data:")
		// fmt.Println(r.SMS)
		return nil
	}
}

func (r *ResultSetT) getAndSortMMS() error { // функция фильтрации данных системы MMS
	mmsData, statusCode, err := mmsCountryRepo.GetMmsData(mmsUrlAddr) // получаем данные из системы MMS
	if statusCode == 200 && err == nil {
		var mmsDSetCountry = make([]MMSData, len(mmsData)) // создаем слайс типа MMSData
		for i, v := range mmsData {                        // проходим по слайсу данных системы MMS
			v.Country = mmsCountryRepo.CountryByCode[countries.Code(v.Country)].Name // в каждом элементе заменяем значение поля кода страны на полное название страны из хранилища MmsCountryRepository
			mmsDSetCountry[i] = MMSData(v)                                           // добавляем в слайс mmsDSetCountry каждый обновленный элемент слайса системы MMS, приводя его к типу MMSData
		}
		var providerSortedSlice = make([]MMSData, len(mmsData)) // инициализируем слайс, для дальнейшей сортировки по имени провайдера
		copy(providerSortedSlice, mmsDSetCountry)               // копируем в него элементы из слайса типа MMSData, с полными именами городов
		for i := 0; i < len(providerSortedSlice); i++ {         // для сортировки используем "сортировку выбора"
			var minIdx = i
			for j := i; j < len(providerSortedSlice); j++ {
				if providerSortedSlice[j].Provider < providerSortedSlice[minIdx].Provider {
					minIdx = j
				}
			}
			providerSortedSlice[i], providerSortedSlice[minIdx] = providerSortedSlice[minIdx], providerSortedSlice[i]
		}
		var countrySortedSlice = make([]MMSData, len(mmsData)) // инициализируем слайс, для дальнейшей сортировки по имени страны
		copy(countrySortedSlice, mmsDSetCountry)
		for i := 0; i < len(countrySortedSlice); i++ {
			var minIdx = i
			for j := i; j < len(countrySortedSlice); j++ {
				if countrySortedSlice[j].Country < countrySortedSlice[minIdx].Country {
					minIdx = j
				}
			}
			countrySortedSlice[i], countrySortedSlice[minIdx] = countrySortedSlice[minIdx], countrySortedSlice[i]
		}
		r.MMS = [][]MMSData{ // заполняем поле MMS структуры ResultSetT. Инициализируем двумерный слайс и присваиваем значения
			providerSortedSlice,
			countrySortedSlice,
		}
		// fmt.Println("MMS system data:")
		// fmt.Println(r.MMS)
		return nil
	} else if err != nil {
		return err
	} else {
		fmt.Printf("Error receiving data about MMS system: StatusCode %v\n", statusCode)
		return nil
	}
}

func (r *ResultSetT) getAndSortVoice() error { // функция фильтрации данных системы Voicecall
	voiceCallData, err := voiceCountryRepo.GetVoiceData(voiceFileName) // получаем данные из системы VoiceCall
	if err != nil {
		return err
	} else {
		var vcData []VoiceCallData // создаем слайс типа VoiceCallData
		for _, v := range voiceCallData {
			vcData = append(vcData, VoiceCallData(v))
		}
		r.VoiceCall = vcData // заполняем поле VoiceCall структуры ResultSetT. Данные из системы никак не модифицируются, просто передаем их в поле.
		// fmt.Println("VoiceCall system data:")
		// fmt.Println(r.VoiceCall)
		return nil
	}
}

func (r *ResultSetT) getAndSortEmail() error { // функция фильтрации данных системы Email
	emailData, err := emailCountryRepo.GetEmailData(emailFileName) // получаем данные из системы Email
	if err != nil {
		return err
	} else {
		r.Email = make(map[string][][]EmailData) // инициализируем мапу для поля Email структуры ResultSetT
		var emData []EmailData                   // создаем слайс типа EmailData
		for _, v := range emailData {            // проходим по слайсу данных системы Email
			emData = append(emData, EmailData(v)) // добавляем в слайс vcData каждый элемент слайса системы Email, приводя его к типу EmailData
		}
		var emailMapByCountry = make(map[string][]EmailData) // создаем мапу c значениями типа слайса EmailData
		for _, v := range emData {                           // проходим по слайсу emData
			emailMapByCountry[v.Country] = append(emailMapByCountry[v.Country], v) // добавляем в мап по ключу == код alpha-2, каждый элемент слайса emData. Для каждого ключа получаем слайс EmailData
		}
		for i, v := range emailMapByCountry { // Для сортировки итерируемся по ключам мапы (по слайсу с значениями для одной страны)
			for i := 0; i < len(v); i++ {
				var minIdx = i
				for j := i; j < len(v); j++ {
					if v[j].DeliveryTime < v[minIdx].DeliveryTime { // сортируем по времени доставки
						minIdx = j
					}
				}
				v[i], v[minIdx] = v[minIdx], v[i]
			}
			r.Email[i] = [][]EmailData{ // для каждой итерации(ключа мапы)
				v[1:4],       // оставляем 3 элемента с самыми быстрыми провайдерами
				v[len(v)-3:], // оставляем 3 элемента с самыми медленными провайдерами
			}
		}
		//fmt.Println("email system data:")
		//fmt.Println(r.Email)
		return nil
	}
}

func (r *ResultSetT) getAndSortBilling() error { // функция фильтрации данных системы Billing
	billingData, err := billing.GetBillingData(billingFileName)
	if err != nil {
		return err
	} else {
		r.Billing = BillingData(billingData) // заполняем поле Billing структуры ResultSetT. Данные из системы никак не модифицируются, просто передаем их в поле.
		//fmt.Println("billing system data:")
		//fmt.Println(r.Billing)
		return nil
	}
}

func (r *ResultSetT) getAndSortSupport() error { // функция фильтрации данных системы Support
	supportData, statusCode, err := support.GetSupportData(supportUrlAddr)
	if statusCode == 200 && err == nil {
		r.Support = make([]int, 0) // инициализируем слайс для поля Support структуры ResultSetT
		var totalActiveTickets int
		for _, v := range supportData {
			totalActiveTickets += v.ActiveTickets // подсчитываем число открытых тикетов
		}
		waitingTime := (60 / 18) * totalActiveTickets // подсчитываем время ожидания для нового тикета
		var supportLoad int                           // нагрузка
		switch {
		case totalActiveTickets < 9: // низкая загрузка
			supportLoad = 1
		case 9 <= totalActiveTickets && totalActiveTickets <= 16: // средняя загрузка
			supportLoad = 2
		default: // высокая загрузка
			supportLoad = 3
		}
		r.Support = []int{supportLoad, waitingTime} // заполняем поле Support структуры ResultSetT. Присваиваем значения.
		// fmt.Println("Support system data:")
		// fmt.Println(r.Support)
		return nil
	} else if err != nil {
		return err
	} else {
		fmt.Printf("Error receiving data about support system: StatusCode %v\n", statusCode)
		return nil
	}
}

func (r *ResultSetT) getAndSortIncident() error { // функция фильтрации данных системы Incident
	incidentData, statusCode, err := incident.GetIncidentData(incidentUrlAddr)
	if statusCode == 200 && err == nil {
		var incData []IncidentData // создаем слайс типа IncidentData
		for _, v := range incidentData {
			incData = append(incData, IncidentData(v))
		}
		for i := 1; i < len(incData); i++ { // Сортируем по статусу. "Active" д.б вначале списка. Используем сортировку "вставками".
			x := incData[i] // значение, для которого будем искать место.
			j := i
			for j = i; j > 0; j-- { // запускаем цикл, идём влево от i по сортированной части
				if incData[j-1].Status > x.Status { // проверяем нужно ли передвигать элементы
					incData[j] = incData[j-1] // передвигаем элементы вверх
				} else {
					break // прерываем цикл, индекс j для вставки x найден
				}
			}
			incData[j] = x // вставляем "x" в нужный элемент с индеком j
		}
		r.Incidents = incData
		// fmt.Println("Incident system data:")
		// fmt.Println(r.Incidents)
		return nil
	} else if err != nil {
		return err
	} else {
		fmt.Printf("Error receiving data about incident system: StatusCode %v\n", statusCode)
		return nil
	}
}

func getResultData() (ResultSetT, error) { // функция получения родительской структуры ResultSetT с отфильтрованными данными всех систем
	var rSetT ResultSetT // создаем структуру типа ResultSetT
	if err := rSetT.getAndSortSMS(); err != nil {
		fmt.Printf("Error receiving data about SMS system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortMMS(); err != nil {
		fmt.Printf("Error receiving data about MMS system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortVoice(); err != nil {
		fmt.Printf("Error receiving data about voiceCall system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortEmail(); err != nil {
		fmt.Printf("Error receiving data about Email system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortBilling(); err != nil {
		fmt.Printf("Error receiving data about billing system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortSupport(); err != nil {
		fmt.Printf("Error receiving data about support system: %v\n", err)
		return rSetT, err
	}
	if err := rSetT.getAndSortIncident(); err != nil {
		fmt.Printf("Error receiving data about incident system: %v\n", err)
		return rSetT, err
	}
	return rSetT, nil
}

func getResultT() ResultT { // функция получения конечной родительской структуры ResultT
	rSetT, err := getResultData() // вызываем функцию получения структуры ResultSetT
	if err != nil {
		rT := ResultT{false, rSetT, err.Error()}
		return rT
	}
	rT := ResultT{true, rSetT, ""}
	return rT
}
