package countries

import (
	"fmt"
	"os"
	"strings"
)

type Code string

type country struct {
	Name   string
	Alpha2 string
}

type CountryRepository struct {
	CountryByCode map[Code]*country
}

func ISOCountryRepository() CountryRepository {
	fileName := "internal/countries/ISOCountries.csv"
	countryMap := CountryRepository{
		CountryByCode: make(map[Code]*country),
	}
	countryMap.fileDataTake(fileName)
	return countryMap
}

func (r *CountryRepository) fileDataTake(fileName string) {
	fileString := string(openFile(fileName))
	fileStringSlice := strings.Split(fileString, "\n")
	for _, v := range fileStringSlice {
		var country country
		workString := strings.TrimSpace(v)
		workStringSplit := strings.Split(workString, ";")
		country.Name = workStringSplit[0]
		country.Alpha2 = workStringSplit[1]
		r.CountryByCode[Code(country.Alpha2)] = &country
	}
}

func openFile(fileName string) []byte {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	buf := make([]byte, fi.Size())
	if _, err = file.Read(buf); err != nil {
		fmt.Println("File reading error:", err)
		panic(err)
	}
	return buf
}
