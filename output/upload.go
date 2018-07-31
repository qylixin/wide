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

package output

import (
	"encoding/json"
	"fmt"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/remote"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

// UploadHandler handles request of uploading.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)
	user := conf.GetUser(username)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	filePath := args["file"].(string)

	if util.Go.IsAPI(filePath) || !session.CanAccess(username, filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	curDir := filepath.Dir(filePath)

	fout, err := os.Create(filePath)

	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	goBuildArgs := []string{}
	goBuildArgs = append(goBuildArgs, "build")
	goBuildArgs = append(goBuildArgs, user.BuildArgs(runtime.GOOS)...)

	cmd := exec.Command("go", goBuildArgs...)
	cmd.Dir = curDir

	fmt.Println("******************************")
	fmt.Println(cmd.Dir)
	fmt.Println("******************************")

	setCmdEnv(cmd, username)
	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("Build project failed")
		logger.Error(err)
		result.Msg = "Build project failed"
		result.Succ = false

		return
	}

	parentDir := filepath.Dir(curDir)
	project := filepath.Base(curDir)
	fmt.Println(project)

	cmd = exec.Command("tar", "-zcvf", project+".tar.gz", project)
	cmd.Dir = parentDir

	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("Compress project failed")
		logger.Error(err)
		result.Msg = "Compress project failed"
		result.Succ = false

		return
	}

	compressPkg := filepath.Join(parentDir, project+".tar.gz")
	result.Msg = compressPkg

	fmt.Println("compress success : ", compressPkg)
	httpSession.Values["chaincode"] = compressPkg
}

// GetChannels handles request of get channel list.
func GetChannels(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*******************  GetChannels  ********************")
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)
	token := httpSession.Values["token"].(string)

	netuuid := r.URL.Query().Get("netuuid")

	channels, err := remote.GetChannel(netuuid, username, token)
	if err != nil {
		logger.Error(err)
		result.Succ = false
		result.Msg = "Get channel list failed."
	}

	result.Data = channels
}

// GetChaincodes handles request of get chaincode list.
func GetChaincodes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*******************  GetChaincodes  ********************")
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)
	token := httpSession.Values["token"].(string)

	netuuid := r.URL.Query().Get("netuuid")

	channels, err := remote.GetChaincode(netuuid, username, token)
	if err != nil {
		logger.Error(err)
		result.Succ = false
		result.Msg = "Get channel list failed."
	}

	result.Data = channels
}

// InstallChaincode handles request of install chaincode.
func InstallChaincode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*******************  InstallChaincode  ********************")
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)
	token := httpSession.Values["token"].(string)

	var args map[string]string

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	fmt.Println("***********  InstallChaincode body  ************")
	fmt.Println(args)

	path := args["path"]
	name := args["name"]
	ccid := args["ccid"]
	channeluuid := args["channeluuid"]
	netuuid := args["netuuid"]

	var err error
	cc := &remote.ResponseInfo{}

	if ccid == "" {
		cc, err = remote.InstallChaincode(netuuid, channeluuid, path, name, username, token)

	} else {
		cc, err = remote.UpgradeChaincode(netuuid, ccid, path, username, token)
	}

	if err != nil {
		logger.Error(err)
		result.Succ = false
		result.Msg = cc.ErrMsg
		result.Code = strconv.Itoa(cc.ErrCode)
		return
	}

	result.Data = cc.Data
}
