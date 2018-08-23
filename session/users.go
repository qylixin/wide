// Copyright (c) 2014-2018, b3log.org & hacpai.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"os/exec"

	"fmt"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
)

const (
	// TODO: i18n

	userExists       = "user exists"
	emailExists      = "email exists"
	userCreated      = "user created"
	userCreateError  = "user create error"
	notAllowRegister = "not allow register"
)

// Exclusive lock for adding user.
var addUserMutex sync.Mutex

// PreferenceHandler handles request of preference page.
func PreferenceHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := HTTPSession.Get(r, "wide-session")

	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	user := conf.GetUser(username)

	if "GET" == r.Method {
		tmpLinux := user.GoBuildArgsForLinux
		tmpWindows := user.GoBuildArgsForWindows
		tmpDarwin := user.GoBuildArgsForDarwin

		user.GoBuildArgsForLinux = strings.Replace(user.GoBuildArgsForLinux, `"`, `&quot;`, -1)
		user.GoBuildArgsForWindows = strings.Replace(user.GoBuildArgsForWindows, `"`, `&quot;`, -1)
		user.GoBuildArgsForDarwin = strings.Replace(user.GoBuildArgsForDarwin, `"`, `&quot;`, -1)

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(user.Locale), "user": user,
			"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version(),
			"locales": i18n.GetLocalesNames(), "gofmts": util.Go.GetGoFormats(),
			"themes": conf.GetThemes(), "editorThemes": conf.GetEditorThemes()}

		t, err := template.ParseFiles("views/preference.html")

		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), 500)

			user.GoBuildArgsForLinux = tmpLinux
			user.GoBuildArgsForWindows = tmpWindows
			user.GoBuildArgsForDarwin = tmpDarwin
			return
		}

		t.Execute(w, model)

		user.GoBuildArgsForLinux = tmpLinux
		user.GoBuildArgsForWindows = tmpWindows
		user.GoBuildArgsForDarwin = tmpDarwin

		return
	}

	// non-GET request as save request

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	args := struct {
		FontFamily            string
		FontSize              string
		GoFmt                 string
		GoBuildArgsForLinux   string
		GoBuildArgsForWindows string
		GoBuildArgsForDarwin  string
		Keymap                string
		Workspace             string
		Username              string
		Password              string
		Email                 string
		Locale                string
		Theme                 string
		EditorFontFamily      string
		EditorFontSize        string
		EditorLineHeight      string
		EditorTheme           string
		EditorTabSize         string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	user.FontFamily = args.FontFamily
	user.FontSize = args.FontSize
	user.GoFormat = args.GoFmt
	user.GoBuildArgsForLinux = args.GoBuildArgsForLinux
	user.GoBuildArgsForWindows = args.GoBuildArgsForWindows
	user.GoBuildArgsForDarwin = args.GoBuildArgsForDarwin
	user.Keymap = args.Keymap
	// XXX: disallow change workspace at present
	// user.Workspace = args.Workspace
	if user.Password != args.Password {
		user.Password = conf.Salt(args.Password, user.Salt)
	}
	user.Email = args.Email

	hash := md5.New()
	hash.Write([]byte(user.Email))
	user.Gravatar = hex.EncodeToString(hash.Sum(nil))

	user.Locale = args.Locale
	user.Theme = args.Theme
	user.Editor.FontFamily = args.EditorFontFamily
	user.Editor.FontSize = args.EditorFontSize
	user.Editor.LineHeight = args.EditorLineHeight
	user.Editor.Theme = args.EditorTheme
	user.Editor.TabSize = args.EditorTabSize

	conf.UpdateCustomizedConf(username)

	now := time.Now().UnixNano()
	user.Lived = now
	user.Updated = now

	result.Succ = user.Save()
}

// LoginHandler handles request of user login.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		// show the login page
		netuuid := r.URL.Query().Get("netuuid")
		username := r.URL.Query().Get("username")
		token := r.URL.Query().Get("token")
		ccid := r.URL.Query().Get("ccid")
		email := r.URL.Query().Get("email")
		

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(conf.Wide.Locale),
			"locale": conf.Wide.Locale, "ver": conf.WideVersion, "year": time.Now().Year(), 
			"netuuid":netuuid,"username":username, "token":token,"ccid":ccid,"email":email}

		t, err := template.ParseFiles("views/login.html")

		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept-Encoding, Content-Encoding, user, pwd")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Content-Type", "application/json")

	// non-GET request as login request
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	args := struct {
		Username  string
		Password  string
		NetworkID string
		CCID      string
		Token     string
	}{}

	args.Username = r.FormValue("username")
	args.Password = r.FormValue("password")
	args.NetworkID = r.FormValue("networkid")
	args.CCID = r.FormValue("ccid")
	args.Token = r.FormValue("token")

	result.Succ = false
	for _, user := range conf.Users {
		if user.Name == args.Username && user.Password == conf.Salt(args.Password, user.Salt) {
			result.Succ = true

			break
		}
	}

	if !result.Succ {
		return
	}

	// create a HTTP session
	httpSession, _ := HTTPSession.Get(r, "wide-session")
	httpSession.Values["username"] = args.Username
	httpSession.Values["token"] = args.Token
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	logger.Debugf("Created a HTTP session [%s] for user [%s] in network [%s]", httpSession.Values["id"].(string), args.Username, args.NetworkID)
}

// LogoutHandler handles request of user logout (exit).
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := HTTPSession.Get(r, "wide-session")

	httpSession.Options.MaxAge = -1
	httpSession.Save(r, w)
}

// SignUpUserHandler handles request of registering user.
func SignUpUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept-Encoding, Content-Encoding, user, pwd")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Content-Type", "application/json")
	if "GET" == r.Method {
		// show the user sign up page

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(conf.Wide.Locale),
			"locale": conf.Wide.Locale, "ver": conf.WideVersion, "dir": conf.Wide.UsersWorkspaces,
			"pathSeparator": conf.PathSeparator, "year": time.Now().Year()}

		t, err := template.ParseFiles("views/sign_up.html")

		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}

	// non-GET request as add user request

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	fmt.Println("***********************SignUpUserHandler***********************")
	//if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
	//	logger.Error(err)
	//	result.Succ = false
	//	body, err := ioutil.ReadAll(r.Body)
	//	fmt.Println("body : ", string(body))
	//	fmt.Println("read body err : ", err)
	//	return
	//}

	fmt.Println("username : ", username)
	fmt.Println("password : ", password)
	fmt.Println("email : ", email)

	msg := addUser(username, password, email)
	fmt.Println("msg : ", msg)
	if userCreated != msg {
		result.Succ = false
		result.Msg = msg

		return
	}

	// create a HTTP session
	httpSession, _ := HTTPSession.Get(r, "wide-session")
	httpSession.Values["username"] = username
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)
}

// FixedTimeSave saves online users' configurations periodically (1 minute).
//
// Main goal of this function is to save user session content, for restoring session content while user open Wide next time.
func FixedTimeSave() {
	go func() {
		defer util.Recover()

		for _ = range time.Tick(time.Minute) {
			SaveOnlineUsers()
		}
	}()
}

// CanAccess determines whether the user specified by the given username can access the specified path.
func CanAccess(username, path string) bool {
	path = filepath.FromSlash(path)

	userWorkspace := conf.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)

	for _, workspace := range workspaces {
		if strings.HasPrefix(path, workspace) {
			return true
		}
	}

	return false
}

// SaveOnlineUsers saves online users' configurations at once.
func SaveOnlineUsers() {
	users := getOnlineUsers()
	for _, u := range users {
		if u.Save() {
			logger.Tracef("Saved online user [%s]'s configurations", u.Name)
		}
	}
}

func getOnlineUsers() []*conf.User {
	ret := []*conf.User{}

	usernames := map[string]string{} // distinct username
	for _, s := range WideSessions {
		usernames[s.Username] = s.Username
	}

	for _, username := range usernames {
		u := conf.GetUser(username)

		if "playground" == username { // user [playground] is a reserved mock user
			continue
		}

		if nil == u {
			logger.Warnf("Not found user [%s]", username)

			continue
		}

		ret = append(ret, u)
	}

	return ret
}

// addUser add a user with the specified username, password and email.
//
//  1. create the user's workspace
//  2. generate 'Hello, 世界' demo code in the workspace (a console version and a HTTP version)
//  3. update the user customized configurations, such as style.css
//  4. serve files of the user's workspace via HTTP
//
// Note: user [playground] is a reserved mock user
func addUser(username, password, email string) string {
	if !conf.Wide.AllowRegister {
		return notAllowRegister
	}

	if "playground" == username {
		return userExists
	}

	addUserMutex.Lock()
	defer addUserMutex.Unlock()

	for _, user := range conf.Users {
		if strings.ToLower(user.Name) == strings.ToLower(username) {
			return userExists
		}

		if strings.ToLower(user.Email) == strings.ToLower(email) {
			return emailExists
		}
	}

	workspace := filepath.Join(conf.Wide.UsersWorkspaces, username)

	newUser := conf.NewUser(username, password, email, workspace)
	conf.Users = append(conf.Users, newUser)

	if !newUser.Save() {
		return userCreateError
	}

	conf.CreateWorkspaceDir(workspace)
	helloWorld(workspace)
	chaincode(workspace)
	conf.UpdateCustomizedConf(username)

	http.Handle("/workspace/"+username+"/",
		http.StripPrefix("/workspace/"+username+"/", http.FileServer(http.Dir(newUser.WorkspacePath()))))

	logger.Infof("Created a user [%s]", username)

	return userCreated
}

// helloWorld generates the 'Hello, 世界' source code.
//  1. src/hello/main.go
//  2. src/web/main.go
func helloWorld(workspace string) {
	consoleHello(workspace)
	//webHello(workspace)
}

func chaincode(workspace string) {
	exampleccdir := conf.ConfDir + conf.PathSeparator + "examplecc"
	myccdir := conf.ConfDir + conf.PathSeparator + "mycc"

	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		return
	}

	exsit := func(dir string) bool {
		if _, err := os.Stat(dir); err != nil {
			if os.IsExist(err) {
				return true
			}
			return false
		}
		return true
	}

	if exsit(exampleccdir) {
		cmd := exec.Command("cp", "-r", exampleccdir, dir)
		err := cmd.Run()
		if err != nil {
			logger.Error(err)
		}
	} else {
		logger.Error(exampleccdir + " is not exsit")
	}

	if exsit(myccdir) {
		cmd := exec.Command("cp", "-r", myccdir, dir)
		err := cmd.Run()
		if err != nil {
			logger.Error(err)
		}
	} else {
		logger.Error(myccdir + " is not exsit")
	}

}

func consoleHello(workspace string) {
	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator + "hello"
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		return
	}

	fout, err := os.Create(dir + conf.PathSeparator + "main.go")
	if nil != err {
		logger.Error(err)

		return
	}

	fout.WriteString(conf.HelloWorld)

	fout.Close()
}

func webHello(workspace string) {
	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator + "web"
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		return
	}

	fout, err := os.Create(dir + conf.PathSeparator + "main.go")
	if nil != err {
		logger.Error(err)

		return
	}

	code := `package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, 世界"))
	})

	port := getPort()

	// you may need to change the address
	fmt.Println("Open https://wide.b3log.org:" + port + " in your browser to see the result") 

	if err := http.ListenAndServe(":"+port, nil); nil != err {
		fmt.Println(err)
	}
}

func getPort() string {
	rand.Seed(time.Now().UnixNano())

	return strconv.Itoa(7000 + rand.Intn(8000-7000))
}

`

	fout.WriteString(code)

	fout.Close()
}
