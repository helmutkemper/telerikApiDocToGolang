package main

import (
	"encoding/json"
	"fmt"
	"github.com/dsoprea/go-logging"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type subFileList struct {
	Url  string
	Name string
}

type Block struct {
	Id      string
	All     string
	Content string
}

var fileToDownload = []string{
	"https://docs.telerik.com/kendo-ui/api/javascript/ui/autocomplete",
	//"https://docs.telerik.com/kendo-ui/api/javascript/ui/grid",
	//"https://docs.telerik.com/kendo-ui/api/javascript/ui/listbox",
}

var fileDownloadStructName = "kendoAutoComplete"

func main() {
	if _, err := os.Stat("./data.json"); os.IsNotExist(err) {
		download()
	}

	process()
}

func process() {

	file, err := ioutil.ReadFile("./data.json")
	if err != nil {
		fmt.Printf("ioutil.ReadFile.error: %v", err.Error())
		os.Exit(1)
	}

	var dataToCode = make(map[string]interface{})

	err = json.Unmarshal(file, &dataToCode)
	if err != nil {
		fmt.Printf("ioutil.Unmarshal.error: %v", err.Error())
		os.Exit(1)
	}

	structList := make(map[string]interface{})

	for id, data := range dataToCode {
		idList := strings.Split(id, ".")
		var key, element string
		if len(idList) > 1 {
			key = idList[len(idList)-2]
			element = idList[len(idList)-1]
		} else {
			key = idList[len(idList)-1]
			element = "main"
		}

		if structList[key] == nil {
			structList[key] = make(map[string]interface{})
		}

		structList[key].(map[string]interface{})[element] = data
	}

	var keys []string
	for structName, structKeys := range structList {
		if structKeys.(map[string]interface{})["main"] != nil {
			keys = append(keys, structName)
		}
	}
	sort.Strings(keys)

	for _, structName := range keys {
		data := structList[structName]

		subProcessData(data.(map[string]interface{})["main"], structName, fileDownloadStructName)
	}

	//fmt.Printf("\n\n\n\n\n")

	for _, structName := range keys {
		data := structList[structName]

		if len(data.(map[string]interface{})) == 1 {
			continue
		}

		//fmt.Printf("type %v struct {\n", strings.Title(structName))
		for subStructName, subStructData := range data.(map[string]interface{}) {
			if subStructName == "main" {
				continue
			}

			subProcessData(subStructData, subStructName, fileDownloadStructName)
		}
		//fmt.Printf("}\n\n\n\n\n")
	}

	keys = []string{}
	for structName, structKeys := range structList {
		if structKeys.(map[string]interface{})["main"] == nil {
			keys = append(keys, structName)
		}
	}
	sort.Strings(keys)

	for _, structName := range keys {
		data := structList[structName]

		//fmt.Printf("type %v struct {\n", strings.Title(structName))
		for subStructName, subStructData := range data.(map[string]interface{}) {
			subProcessData(subStructData, subStructName, fileDownloadStructName)
		}
		//fmt.Printf("}\n\n\n\n\n")
	}

	//fmt.Println("fim!")
}

func subProcessData(data interface{}, structName, prefix string) {
	return
	defaultValue := data.(map[string]interface{})["default"].(string)

	description := data.(map[string]interface{})["description"].(string)

	//<a href="/kendo-ui/api/javascript/kendo/methods/template">template</a>
	re := regexp.MustCompile(`(<a href="(/kendo-ui/api/javascript/.*?)">(.*?)</a>)`)
	replaceLink := re.FindAllStringSubmatch(description, -1)
	for _, replaceList := range replaceLink {
		description = strings.Replace(description, replaceList[0], fmt.Sprintf("%v [ https://docs.telerik.com%v ]", replaceList[3], replaceList[2]), -1)
	}

	description = strings.Replace(description, "&lt;", "<", -1)
	description = strings.Replace(description, "&gt;", ">", -1)
	description = strings.Replace(description, "<code>", "`", -1)
	description = strings.Replace(description, "</code>", "´", -1)
	description = strings.Replace(description, "<blockquote>", "Important:", -1)
	description = strings.Replace(description, "</blockquote>", "", -1)
	description = strings.Trim(description, "")

	if defaultValue != "" {
		description += fmt.Sprintf(" (default: %v)", defaultValue)
	}

	jsType := ""
	jsTagType := ""
	typePrefix := ""
	//jsType:"*JavaScript,string"
	jsTypes := data.(map[string]interface{})["types"].([]interface{})
	if strings.Contains(structName, "Template") || strings.Contains(structName, "template") {
		jsType = "interface{}"
		jsTagType = " jsType:\"*JavaScript,string\""
	} else if hasType(jsTypes, "Function") == true {
		jsType = "*JavaScript"
	} else if hasType(jsTypes, "kendo.data.DataSource") == true {
		jsType = "interface{}"
		jsTagType = " jsType:\"*KendoDataSource,string,*map[string]interface {},[]string\""
	} else if hasType(jsTypes, "String") == true && hasType(jsTypes, "Boolean") == true && len(jsTypes) == 2 {
		jsType = "interface{}"
		jsTagType = " jsType:\"Boolean,string\""
	} else if hasType(jsTypes, "String") == true && hasType(jsTypes, "Number") == true && len(jsTypes) == 2 {
		jsType = "int"
	} else if hasType(jsTypes, "String") == true && len(jsTypes) == 1 {
		jsType = "string"
	} else if hasType(jsTypes, "Date") == true && len(jsTypes) == 1 {
		jsType = "time.Time"
	} else if hasType(jsTypes, "Boolean") == true && len(jsTypes) == 1 {
		jsType = "Boolean"
	} else if hasType(jsTypes, "Array") == true && len(jsTypes) == 1 {
		jsType = "[]fixme"
	} else if hasType(jsTypes, "Object") == true && len(jsTypes) == 1 {
		//typePrefix = prefix
		jsType = "*" + prefix + strings.Title(structName)
	} else if hasType(jsTypes, "Number") == true && len(jsTypes) == 1 {
		jsType = "int"
	} else if hasType(jsTypes, "Boolean") == true && hasType(jsTypes, "Object") == true && len(jsTypes) == 2 {
		//typePrefix = prefix
		jsTagType = " jsType:\"Boolean,*" + prefix + strings.Title(structName) + "\""
		jsType = "interface{}"
	} else {
		jsType = ""
	}

	examples := data.(map[string]interface{})["examples"].([]interface{})

	fmt.Printf("/*\n")
	fmt.Printf("@see %v/configuration/%v\n\n", data.(map[string]interface{})["see"], strings.ToLower(structName))
	fmt.Printf("%v\n", description)
	fmt.Printf("*/\n")
	for _, example := range examples {
		fmt.Printf("%v\n", example)
	}

	fmt.Printf("%v %v%v `jsObject:\"%v\"%v`\n\n", strings.Title(structName), typePrefix, jsType, structName, jsTagType)
}

func hasType(data []interface{}, jsType string) bool {
	for _, v := range data {
		if v.(string) == jsType {
			return true
		}
	}

	return false
}

func download() {

	f, err := os.OpenFile("./toTag.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var dataToCode = make(map[string]interface{})
	var dataToView = make([]string, 0)

	for _, pageToDownload := range fileToDownload {
		file := DownloadFile(pageToDownload)

		subListToDownload := filterSubFiles(file)
		for _, newFileToDownload := range subListToDownload {
			//fmt.Println("https://docs.telerik.com/kendo-ui/api/javascript/ui/" + newFileToDownload.Url)
			file := DownloadFile("https://docs.telerik.com/kendo-ui/api/javascript/ui/" + newFileToDownload.Url)

			file = strings.Replace(file, "<h3", "<!-- gambiarra --><h3", -1)
			file = strings.Replace(file, "</article>", "<!-- gambiarra --></article>", -1)

			for _, blockH3 := range getBlockOfDescription(file) {
				if !strings.Contains(blockH3.Id, ".") {
					var pass = true
					for _, v := range dataToView {
						if v == blockH3.Id {
							pass = false
							break
						}
					}

					if pass == true {
						dataToView = append(dataToView, blockH3.Id)
					}
				}

				dataToCode[blockH3.Id] = make(map[string]interface{})

				h3Content := getH3Content(blockH3.All)

				dataToCode[blockH3.Id].(map[string]interface{})["see"] = pageToDownload
				dataToCode[blockH3.Id].(map[string]interface{})["description"] = getDescription(blockH3.All)
				dataToCode[blockH3.Id].(map[string]interface{})["default"] = getDefaultValue(h3Content)
				dataToCode[blockH3.Id].(map[string]interface{})["types"] = getTypes(h3Content)
				dataToCode[blockH3.Id].(map[string]interface{})["examples"] = getExamples(blockH3.Content)

				dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = pageToDownload + "\n\n"

				types := getTypes(h3Content)
				defaultValue := getDefaultValue(h3Content)
				if len(types) != 0 && defaultValue != "" {

					dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string) + "Type: " + strings.Join(types, ", ") + " (default: " + defaultValue + ")\n\n"

				} else if len(types) == 0 && defaultValue != "" {

					dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string) + "(default: " + defaultValue + ")\n\n"

				} else if len(types) != 0 && defaultValue == "" {

					dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string) + "Type: " + strings.Join(types, ", ") + "\n\n"

				}

				dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string) + "\n\n" + strings.Join(getExamples(blockH3.Content), "\n")

				dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = pageToDownload + "\n\n" + "Default: " + getDefaultValue(h3Content) + "\n" + "Type: " + strings.Join(getTypes(h3Content), ", ") + "\n\n" + "Description: " + getDescription(blockH3.All) + "\n\n" + strings.Join(getExamples(blockH3.Content), "\n")
				dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = strings.Replace(dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string), "\n", "\\n", -1)
				dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = strings.Replace(dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string), "\"", "\\\"", -1)
				//dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = "jsonSchema_description:\"" + dataToCode[blockH3.Id].(map[string]interface{})["toTag"].(string) + "\""

				// jsonSchema_complex

				// fixme: apagar - início
				// dataToCode[blockH3.Id].(map[string]interface{})["toTag"] = "ola mundo"
				// fixme: apagar - fim

				/*
					if _, err = f.WriteString(fmt.Sprintf("%v %v\n", blockH3.Id, "")); err != nil {
						panic(err)
					}
				*/
			}
		}
	}

	var keyMax = 0
	for k := range dataToCode {
		keyCount := len(strings.Split(k, "."))

		if keyCount > keyMax {
			keyMax = keyCount
		}
	}

	var schema = make(map[string]interface{})
	dotCount := keyMax

	for k, v := range dataToCode {
		schema[k] = make(map[string]interface{})
		schema[k].(map[string]interface{})["description"] = v.(map[string]interface{})["toTag"]
		schema[k].(map[string]interface{})["type"] = make(map[string]interface{})

		for _, jsType := range v.(map[string]interface{})["types"].([]string) {
			jsType = strings.ToLower(jsType)

			// fixme: apagar - início
			//        isto foi colocado para o caso de objetos tipo anime e deve ser removido
			//        quando o projeto estiver acabado

			//if len(v.(map[string]interface{})["types"].([]string)) > 1 && jsType == "boolean" {
			//  continue
			//}

			// fixme: apagar - fim

			schema[k].(map[string]interface{})["types"] = make(map[string]interface{})
			schema[k].(map[string]interface{})["types"].(map[string]interface{})[jsType] = make(map[string]interface{})

			schema[k].(map[string]interface{})["properties"] = make(map[string]interface{})
		}
	}

	dotCount = keyMax + 1
	for {
		dotCount--
		if dotCount <= 1 {
			break
		}

		var dataToDelete = make([]string, 0)

		for keyInProcess, v := range schema {

			keyCount := len(strings.Split(keyInProcess, "."))
			if keyCount != dotCount {
				continue
			}

			for keyToFind := range schema {

				if strings.LastIndex(keyInProcess+".", keyToFind) == 0 {
					keyNew := strings.Replace(keyInProcess, keyToFind+".", "", -1)
					if strings.LastIndex(keyNew, ".") != -1 {
						continue
					}

					dataToDelete = append(dataToDelete, keyInProcess)

					schema[keyToFind].(map[string]interface{})["properties"].(map[string]interface{})[keyNew] = v

					fmt.Printf("entrou: %v\n", keyNew)
				}

			}
		}

		for _, v := range dataToDelete {
			//delete( dataToCode, v )
			delete(schema, v)
		}
	}

	js, err := json.Marshal(schema)
	if err != nil {
		log.Panic(err.Error())
	}
	fmt.Printf("%s", js)

	for {
		pass := false

		for k, v := range dataToCode {
			tagKeysFound := strings.Split(k, ".")

			if len(tagKeysFound) != dotCount {
				pass = true
				continue
			}

			tagKey := tagKeysFound[dotCount-1]

			if dotCount == 1 {
				schema[tagKey] = make(map[string]interface{})
				schema[tagKey].(map[string]interface{})["description"] = v.(map[string]interface{})["toTag"]
				schema[tagKey].(map[string]interface{})["type"] = make(map[string]interface{})

				for _, jsType := range v.(map[string]interface{})["types"].([]string) {
					jsType = strings.ToLower(jsType)

					// fixme: apagar - início
					//        isto foi colocado para o caso de objetos tipo anime e deve ser removido
					//        quando o projeto estiver acabado

					if len(v.(map[string]interface{})["types"].([]string)) > 1 && jsType == "boolean" {
						continue
					}

					// fixme: apagar - fim

					schema[tagKey].(map[string]interface{})["properties"] = make(map[string]interface{})

					schema[tagKey].(map[string]interface{})["types"] = make(map[string]interface{})
					schema[tagKey].(map[string]interface{})["types"].(map[string]interface{})[jsType] = make(map[string]interface{})
				}
			}
		}

		dotCount += 1

		if pass == false {
			break
		}
	}

	fileToJSon, err := json.Marshal(dataToCode)
	if err != nil {
		fmt.Printf("json.Marshal.error: %v", err.Error())
		os.Exit(1)
	}

	ioutil.WriteFile("./data.json", fileToJSon, 0664)
}

func getDescription(h3All string) string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:</h3>\s*(.*?)\s*<h4>)`)
	allTypes := regexpLiMainPage.FindAllStringSubmatch(h3All, -1)

	for _, typeLine := range allTypes {
		typeLine[1] = strings.Replace(typeLine[1], "<p>", "", -1)
		typeLine[1] = strings.Replace(typeLine[1], "</p>", "", -1)
		return typeLine[1]
	}

	return ""
}

func getTypes(h3Content string) []string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:<code>(.*?) *\|*</code>)`)
	allTypes := regexpLiMainPage.FindAllStringSubmatch(h3Content, -1)

	ret := make([]string, len(allTypes))

	for k, typeLine := range allTypes {
		ret[k] = typeLine[1]
	}

	return ret
}

func getDefaultValue(h3Content string) string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:<em>\(default:\s*(.*?)\)</em>)`)
	allTypes := regexpLiMainPage.FindAllStringSubmatch(h3Content, -1)

	for _, typeLine := range allTypes {
		return typeLine[1]
	}

	return ""
}

/**
 * Retorna o bloco contido dentro de H3 e o nome da chave Kendo-UI
 * @param {string} file - página html
 * @return {Block}
 *   Block.Id {string}      - id da chave Kendo-UI
 *   Block.Content {string} - conteúdo da tag html <h3>
 */
func getBlockOfDescription(file string) []Block {
	regexpLiMainPage := regexp.MustCompile(`(?smi:<h3\s+id=['"](.*?)['"]>(.*?)(?:<!-- gambiarra -->))`)
	allTypes := regexpLiMainPage.FindAllStringSubmatch(file, -1)

	ret := make([]Block, len(allTypes))

	for key, typeLine := range allTypes {
		ret[key] = Block{
			Id:      typeLine[1],
			All:     typeLine[0],
			Content: typeLine[2],
		}
	}

	return ret
}

func getH3Content(file string) string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:<h3\s+id=['"].*?['"]>(.*?)(?:</h3>))`)
	allTypes := regexpLiMainPage.FindAllStringSubmatch(file, -1)

	for _, typeLine := range allTypes {
		return typeLine[1]
	}

	return ""
}

/*
func getExamples(file string) []string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:(<h4>.*?(?P<DESCRIPTION>Example.*?)</h4>.*?<pre>.*?<code>(?P<CODE>.*?)</code>.*?</pre>))`)
	allExamples := regexpLiMainPage.FindAllStringSubmatch(file, -1)

	ret := make([]string, len(allExamples))

	for k, exampleLine := range allExamples {
		exampleLine[3] = strings.Replace(exampleLine[3], "&lt;", "<", -1)
		exampleLine[3] = strings.Replace(exampleLine[3], "&gt;", ">", -1)
		tmp := strings.Split(exampleLine[3], "\n")
		exampleLine[3] = "  //    " + strings.Join(tmp, "\n  //    ")

		ret[k] = "  //    \n  //    " + exampleLine[2] + "\n  //    \n" + exampleLine[3]
	}

	return ret
}
*/

func getExamples(file string) []string {
	regexpLiMainPage := regexp.MustCompile(`(?smi:(<h4>.*?(?P<DESCRIPTION>Example.*?)</h4>.*?<pre>.*?<code>(?P<CODE>.*?)</code>.*?</pre>))`)
	allExamples := regexpLiMainPage.FindAllStringSubmatch(file, -1)

	ret := make([]string, len(allExamples))

	for k, exampleLine := range allExamples {
		exampleLine[3] = strings.Replace(exampleLine[3], "&lt;", "<", -1)
		exampleLine[3] = strings.Replace(exampleLine[3], "&gt;", ">", -1)

		ret[k] = exampleLine[2] + "\n\n" + exampleLine[3]
	}

	return ret
}

func filterSubFiles(file string) []subFileList {
	subFiles := make([]subFileList, 0)

	//regexpLiMainPage := regexp.MustCompile(`(?ms:<article>(.*?)<h2 id="fields">)`)
	regexpLiMainPage := regexp.MustCompile(`(?ms:<h2 id="configuration">(.*?)</ul>)`)
	allElementsLiMainPage := regexpLiMainPage.FindAllStringSubmatch(file, -1)
	file = allElementsLiMainPage[0][1]

	regexpLiMainPage = regexp.MustCompile(`<li>\s*<a href=['"](?P<URL>.*?)['"]\s*>(?P<DATA_NAME>.*?)</a>\s*</li>`)
	allElementsLiMainPage = regexpLiMainPage.FindAllStringSubmatch(file, -1)
	for _, ElementLiMainPage := range allElementsLiMainPage {
		fileData := subFileList{}
		fileData.Name = ElementLiMainPage[2]
		fileData.Url = ElementLiMainPage[1]

		subFiles = append(subFiles, fileData)
	}

	return subFiles
}

func DownloadFile(url string) string {

	// Create the file
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	return string(body)
}
