package main

import (
	"fmt"

	"bitbucket.org/Soytul/library-go-http/tulhttp"
)

type Logger struct{}

func (l Logger) Debug(...interface{}) {}
func (l Logger) Error(...interface{}) {}
func (l Logger) Info(...interface{})  {}

func main() {
	tulhttp.Init(map[string][]string{}, Logger{})
	request := tulhttp.TulRequest{
		Method: "POST",
		Url:    "https://api.br.soytul.com/product/catalog/v1/products/ean/multiple-ean/",
		Body:   []string{"7891040219247"},
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer eyJraWQiOiJ0dWwtYmFja29mZmljZS1iciIsInR5cCI6IkpXVCIsImFsZyI6IlJTMzg0In0.eyJzdWIiOiIxZDU5ZGY2Mi0wZjI3LTRmZjEtYWY4ZS1mYThhMTU1OTBiY2UiLCJjb3VudHJ5IjoiQlIiLCJwZXJzb25fdHlwZSI6ImJhY2tvZmZpY2UiLCJyb2xlcyI6WyJiYWNrb2ZmaWNlIl0sImlzcyI6InNoYXJlZC1zZWN1cml0eSIsIm5hbWUiOiJhbmRyZXMgY2FycmlsbG8iLCJ0eXAiOiJCZWFyZXIiLCJlbnYiOiJwcm9kIiwiZXhwIjoxNzA1MDk3OTExLCJpYXQiOjE3MDUwOTcwMTEsImVtYWlsIjoiYW5kcmVzLmNhcnJpbGxvQHNveXR1bC5jb20iLCJkZXZpY2VfdXVpZCI6IjBiNWY5ZGQ3LTZlNTAtNGQ2Yy1iNzliLWMzNTQ2NDUxNjBmYSJ9.Do-xWiHV5EAXuoIdPFEMOaKR1MerCXAQnCIjjuSP3UuI1irnXu8oVjLiCDe101y7QRpk23AKeOHa7KtRotAtGaLVkV9n-rqHmlNevt3v-Fco37Jw_B-BAZPLIC0KsElooZQHcBGJh7euQU0SAx9N_PL1rneXSoKfnUaDuxi0mkFxZfL115GAhvR37zecS1EKR8A4HlR_OUfuT1V6HSInQU1c5p1G_knt2947-QqgZYxQCmmC-iJx9oGTQRNY_1Xp6fQ109xbsj2WCW1TogwiukVFc8ZXXaUriU8EPKBg38g2YLeLowYvHbV80OpTVKi-fcqoU3h5_pDjWzzjNG7uTXzgTZk145WEMYtVU6xrWZ2uz_xzWMInk0dFg2UoPxkDQhyU9TL0l_Jg3gFo_UfwFKVEuB1M3KEScnSbbbfeDo1AhpkkdvM7ODJMUuR8WTdW6HpMq9CmfaDbJ7EIdRvq_Nr7bgTNw7s2WIhCRdEVDsk00z2Sc9dHki-yHOi8imoh",
		},
	}

	response := request.Fetch()

	if response.Error != nil {
		fmt.Println(response.Error, response.DataAsMap())
		return
	}

	fmt.Println(response.DataAsSlice())
}
