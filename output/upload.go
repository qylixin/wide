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
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var currentChaincodePkg map[string]string

func init() {
	currentChaincodePkg = make(map[string]string)
}

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

	fmt.Println("compress success : ", compressPkg)
	httpSession.Values["chaincode"] = compressPkg
}
