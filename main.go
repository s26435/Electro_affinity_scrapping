/*
Stworzono:
Jan Wolski
27.02.2025
*/
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"encoding/csv"
	"os"
	r "github.com/go-rod/rod"
)

type dataframe struct{
	name []string
	ea_value []float32
	formula []string
}

func (df *dataframe)add(name string, value float32, cas string){
	df.name = append(df.name, name)
	df.ea_value = append(df.ea_value, value)
	df.formula = append(df.formula, cas)
}

func (df *dataframe) save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("błąd tworzenia pliku: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Związek", "EA (eV)", "smiles"}); err != nil {
		return fmt.Errorf("błąd zapisu nagłówka: %v", err)
	}

	for i := range df.name {
		record := []string{
			df.name[i], 
			fmt.Sprintf("%.6f", df.ea_value[i]), 
			df.formula[i],
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("błąd zapisu wiersza %v: %v", record, err)
		}
	}

	fmt.Printf("Dane zapisane do %s\n", filename)
	return nil
}


func getUrlString(start, end float32) string{
	return fmt.Sprintf("https://webbook.nist.gov/cgi/cbook.cgi?Value=%f%%3B%f&VType=EA&Formula=&AllowExtra=on&Units=SI", start, end)  
}

func getSelectorStrong(nth int)string{
	return fmt.Sprintf("#main > ol > li:nth-child(%d) > strong", nth+1)
}

func getSelectorA(nth int)string{
	return fmt.Sprintf("#main > ol > li:nth-child(%d) > a", nth+1)
}

const chemical_formula_selector string = "#main > ul:nth-child(2) > li:nth-child(1)"

func scrapFormula(i int, page *r.Page)string{
	page.MustElement(getSelectorA(i)).MustClick()
	page.MustWaitLoad()
	ch_f__val := page.MustElement(chemical_formula_selector).MustText()
	index :=  strings.Index(ch_f__val, " ")
		if index != -1 {
			ch_f__val = ch_f__val[index+1:] 
		} else {
			ch_f__val = ""
		}
	if ch_f__val != ""{
		fmt.Println(ch_f__val)
	}
	page.MustNavigateBack()
	return ch_f__val
}

func scrappOne(start, end float32, page *r.Page, df *dataframe){
	var err error
	var th int
	page.MustNavigate(getUrlString(start, end))
	how_many_s, err := page.Timeout(time.Second).Element("#main > p:nth-child(2)")
	if err != nil{
		th = 0 
	}else{
		how_many := how_many_s.MustText()
		index :=  strings.Index(how_many, " ")
		if index != -1 {
			result := how_many[:index] 
			th, err = strconv.Atoi(result)
			if err != nil {
				fmt.Println(result)
				if result == "Due"{
					th=400
				}else{
					th=0
				}
				
			}
		} else {
			panic("brak liczby wyszukań")
		}
	}
	
	for i:=0; i < th; i++ {
		strong, err := page.Timeout(time.Second).Element(getSelectorStrong(i))
		if err != nil{
			return
		}
		value := strong.MustText()
		d, err := page.Element(getSelectorA(i))
		if err != nil {
			return
		}
		name := d.MustText()
		value = value[3:(len(value)-3)]
		float_value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		chemical_formula := scrapFormula(i, page)
		if(chemical_formula != ""){
			df.add(name, float32(float_value), chemical_formula)
		}
	}
}

func ScrappAll(logging bool) dataframe{
	const start float32 = -0.6
	const end float32 = 9.0
	const interval float32 = 0.2
	browser := r.New().MustConnect().NoDefaultDevice()
	defer browser.MustClose()
	page :=  browser.MustPage(getUrlString(0, 1))
	var x float32 = (end - start)/interval
	how_many := int(math.Ceil(float64(x)))
	var df dataframe
	for i:=0; i < how_many; i++{
		s:= start+float32(i)*interval
		e:= start+float32(i+1)*(interval)
		if logging{
			fmt.Printf("Serching interval: %f - %f\n", s,e)
		}
		scrappOne(s,e , page, &df)
	}
	
	return df
}

func main() {
	df := ScrappAll(true)
	fmt.Print(len(df.name))
	df.save("dataset.csv")
}
