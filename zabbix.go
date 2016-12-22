package zabbix

import(
    "fmt"
    "strings"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "bytes"
)
//===========================================================================
type Zabbix struct {
    url  string
    sid  string
    Rpc  string   `json:"jsonrpc"`
    Code string   `json:"result"`
    Err  ZError   `json:"error"`
    Id   int      `json:"id"`
}
//  --
type ZError struct {
    Code    int     `json:"code"`
    Message string  `json:"message"`
    Data    string  `json:"data"`
}
//  --  --  --
type ZHost struct {	// hosts
    Code string `json:"host"`
    Name string `json:"name"`
    Id   string `json:"hostid"`
}
//  --
type ZTrigger struct {
    Description string  `json:"description"`
    Triggerid   string  `json:"triggerid"`
    Priority    string  `json:"priority"`
    Lastchange  string  `json:"lastchange"`
    Status      string  `json:"status"`
    Value       string  `json:"value"`
    Host       []ZHost  `json:"hosts"`
    Event   interface{} `json:"lastEvent"`
    Comments    string  `json:"comments"`
//    Templateid  string  `json:"templateid"`
//    Types       string  `json:"type"`
//    State       string  `json:"state"`
//    Flags       string  `json:"flags"`
//    Error       string  `json:"error"`
}
//  --
type ZHistory struct {
    Id      string    `json:"id"`
    ItemId  string    `json:"itemid"`
    Clock   string    `json:"clock"`
    Value   string    `json:"value"`
    Ns      string    `json:"ns"`
}
//===========================================================================
func ( z *Zabbix ) rcall( qr string ) ( res []byte, err error ){ // Запрос к Zabbix - получаем массив байт
    if z.sid == "" || z.url == "" || qr == "" {
        return res, fmt.Errorf( "zbxRequest: Id:%d, Url: %s. msg: %s\n", z.Id, z.url, qr )
    }
//  --
    qstr := `{"jsonrpc":"2.0",`+ qr
    if z.Code != "" { qstr += `,"auth":"` +z.Code +`"` }
    qstr +=`,"id":`+ z.sid +`}`
//  --
//    fmt.Println( "<<< " + qstr)
//  --
    client := &http.Client{}
    req, err := http.NewRequest( "POST", z.url, bytes.NewBufferString( qstr ))
    if err != nil {
        return res, err
    }
    req.ContentLength = int64( len( qstr ))
    req.Header.Add( "Content-Type", "application/json-rpc" )
    req.Header.Add( "User-Agent", "SMonitor" )
//  --
    resp, err := client.Do( req )
    if err != nil {
        return res, err
    }
    if resp.StatusCode != 200 && resp.ContentLength < 30 {
        return res, fmt.Errorf( "zbxRequest: status:%d, data-len:%d\n", resp.StatusCode, resp.ContentLength )
    }
    res, err = ioutil.ReadAll( resp.Body )
    if err != nil {
        return res, err
    }
    defer resp.Body.Close()
    return res, err
}// End
//  -------------------------------------------------------------------------
func New( url, user, passwd string )( z Zabbix, err error ) { // Инициализация сессии
    if url == "" {
        return z, fmt.Errorf( "ZabbixInit: Error URL" )
    }
    z.sid = "1"
    z.url = "http://"+url+"/zabbix/api_jsonrpc.php"
    zstr, err := z.rcall( `"method":"user.login","params":{"password":"`+ passwd +`","user":"`+ user +`"}` )
    if err != nil {
        return z, err
    }
//  --
//fmt.Printf( "ZabbixInit: %s\n\n", zstr )
//  --
    err = json.Unmarshal( []byte( zstr ), &z )
    if err == nil && z.Rpc == "2.0" && z.Code != "" {
        z.sid  = fmt.Sprint( z.Id )
        return z, err
    }else if err == nil && z.Err.Code != 0 {
        return z, fmt.Errorf( "ZabbixInit: %s %s\n", z.Err.Message, z.Err.Data )
    }
    return z, fmt.Errorf( "ZabbixInit: %s", zstr )
}// End
//  -------------------------------------------------------------------------
func ( z *Zabbix ) Close(){ // Закрытие сессии
    if z.sid == "" || z.url == "" || z.Code == "" {
        return
    }
    zstr, err := z.rcall( `"method":"user.logout","params":[]` )
//fmt.Printf( "zbxSessionClose: %s\n", zstr )
    if err == nil && strings.Contains( string( zstr ), `"result":true` ){
        z.sid = ""
        z.url = ""
        z.Code = ""
        z.Id = 0
//fmt.Printf( "zbxSession - close.\n" )
    }
}// End
//  -------------------------------------------------------------------------

//  -------------------------------------------------------------------------
func ( z *Zabbix ) GetTrigger( gr string )( []*ZTrigger, error ){
    if gr != "" { gr += "," }
    zq := `"method":"trigger.get","params":{"active":"1","only_true":"1","expandData":"1","selectHosts":["host","name"],"selectLastEvent":["eventid"],`+gr+`"output":"extend"}`
//  --
    type qtriggers struct {
        Err ZError      `json:"error"`
        Res []*ZTrigger `json:"result"`
    }
//  --
    trs := qtriggers{}
    zstr, err := z.rcall( zq )

//fmt.Printf( "zbx: %q\n", zstr )

    if err == nil {
        err = json.Unmarshal( zstr, &trs )
        if err == nil && trs.Err.Code != 0 {
            err = fmt.Errorf( "GetTrigger: %s %s\n", trs.Err.Message, trs.Err.Data )
        }
    }
return trs.Res, err
}
//  -------------------------------------------------------------------------
func ( z *Zabbix ) GetHostId( hosts string )( string, error ){
    type zHostid struct {
        Id      string    `json:"hostid"`
        Host    string    `json:"host"`
        Name    string    `json:"name"`
    }
    type qHostid struct {
        Err ZError     `json:"error"`
        Res []*zHostid `json:"result"`
    }
//  --
    qtr := qHostid{}
    zstr, err := z.rcall( `"method":"host.get","params":{"filter":{"host":"`+hosts+`"}}` )
    if err == nil {
        err = json.Unmarshal( zstr, &qtr )
        if err == nil && qtr.Err.Code != 0 {
            err = fmt.Errorf( "GetHostId: %s %s\n", qtr.Err.Message, qtr.Err.Data )
        }
//fmt.Printf( "\nzbxGetHostId: %s\n\n", qtr.Res[0].Id )
    }
    if err == nil && len(qtr.Res) > 0 {
        return qtr.Res[0].Id, nil
    }
return "", err
}
//  -------------------------------------------------------------------------
func ( z *Zabbix ) GetHistory( hosts string )( []*ZHistory, error ){
    type qHistory struct {
        Err ZError      `json:"error"`
        Res []*ZHistory `json:"result"`
    }
//  --
    qtr := qHistory{}
    hid, err := z.GetHostId( hosts )
    if err == nil {
//  --
//fmt.Printf( "HostId: %s\n", hid )
//  --
        zstr, err := z.rcall( `"method":"history.get","params":{"history":4,"limit":80,"sortfield":"clock","sortorder":"DESC","hostids":"`+hid+`"}` )
        if err == nil {
            err = json.Unmarshal( zstr, &qtr )
            if err == nil && qtr.Err.Code != 0 {
                err = fmt.Errorf( "GetHostId: %s %s\n", qtr.Err.Message, qtr.Err.Data )
            }
        }
    }
return qtr.Res, err
}
//===========================================================================
//===========================================================================
