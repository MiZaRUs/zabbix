# GO zabbix_get
Простой модуль получения данных от zabbix.

Пример запросов:
```go
//  -------------------------------------------------------------------------
package main

import(
    "fmt"
    "strconv"
    "time"
    "regexp"
    "zabbix"
)

func main(){

// Фильтрация по хостам
    var invalidHosts = regexp.MustCompile( `((^(servxxx|servyyy))|(((-old|-dbx|-wb).?)$))` )

// Фильтрация по назвынию триггера
    var invalidNames = regexp.MustCompile( `(Zabbix|Ошибка загрузки|Oracle - PGA)` )

// Инициируем подключение
    zbx, err := zabbix.New( "192.168.100.182", "script-user", "test" )
    if err != nil {
        fmt.Printf("ERROR: %s\n", err )
    }else{
        defer zbx.Close()	// с последующим закрытием
        var trAllCnt, trCnt int		// счетчики

// Получаем триггеры
        triggers, err := zbx.GetTrigger( `"maintenance":"false"` )
        if err != nil {
            fmt.Printf( "EROOR: %s\n", err )
        }else{
            fmt.Println()
            for _, tr := range triggers {
                if ( !invalidHosts.MatchString( tr.Host[0].Code ))&&( !invalidNames.MatchString( tr.Description )){
                    i, _ := strconv.Atoi( tr.Lastchange )
                    tms := time.Unix( int64(i), 0).Format("2006-01-02 15:04:05")
                    Eventid := ""
                    switch item := tr.Event.(type) {
                        case map[string]interface{}:  Eventid = fmt.Sprintf( "%s", item["eventid"] )
                    }
                    fmt.Printf( " %-30s %s  %-20s %s (%s)  (%s,%s,%s,%s) %s\n", tr.Host[0].Name, tms, tr.Host[0].Code, tr.Description, tr.Comments, tr.Triggerid, tr.Priority, tr.Status, Eventid, tr.Value )
                    trCnt++
                }
                trAllCnt++
            }
        }
    fmt.Printf( "\nТриггеров получено: %d, выбрано: %d, игнорировано: %d\n\n", trAllCnt, trCnt, (trAllCnt - trCnt) )
//  --

// Получаем историю по указанному хосту
        hist, err := zbx.GetHistory( "abcd-db" )
        if err != nil {
            fmt.Printf( "ERROR: %s\n", err )
        }else{
            fmt.Println()
            for _, tr := range hist {
                fmt.Printf( " %s, %s, %s, %s, %s\n", tr.Id, tr.ItemId, tr.Clock, tr.Value, tr.Ns )
            }
        }
        fmt.Println()
//  --
    }
}// end
//  -------------------------------------------------------------------------
