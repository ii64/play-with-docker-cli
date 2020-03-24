package api

/*
get PWD_URL
-stream: fetch instances status using ws

*/

import (
	"os"
	"fmt"
	"log"
	"flag"
	"sync"
	"strings"
	"io/ioutil"
	"path/filepath"
)

var _ = fmt.Printf
var _ sync.Once

type console struct {
	pwd_url string
	dashboard *dashboard
}
func NewConsole() *console {
	return &console{
	}
}
func (c *console) Log(s string, f ...interface{}) {
	log.Printf("[C]: "+s, f...)
}
func (c *console) FSTreeList(params map[string]interface{}) {
	ins, _ := params["inode"];
	if prm, ok := ins.([]string); ok && len(prm) > 0 {
		existIns := c.dashboard.GetInstances()
		collectMatch := map[string]string{}
		aliasMatch := []string{}
		for _, insIdAlias := range prm {
			// check to ins list
			for k, v := range existIns {
				if strings.Contains(k, insIdAlias) {
					aliasMatch   = append(aliasMatch, insIdAlias)
					collectMatch[k] = v.Name
				}
			}
		}
		filePath := os.Args[len(os.Args)-1]
		if len(filePath) == 0 {
			fmt.Printf("error: need file path")
			return
		}
		if len(collectMatch) == 0 {
			fmt.Printf("error: none instance selected\n")
		}
		showWholeTree := false
		for _, iId := range aliasMatch {
			if iId == filePath {
				showWholeTree = true
				break
			}
		}
		for simplifyId, iId := range collectMatch {
			rfs, err := c.dashboard.FSTree(iId)
			if err != nil {
				fmt.Printf("%s: error: %s\n", simplifyId, err)
				continue
			}
			if showWholeTree {
				fmt.Printf("%s\n", rfs.ToString(simplifyId))
			}else{
				dir, err := rfs.GetDir(filePath)
				if err != nil {
					fmt.Printf("%s: error: %s\n", simplifyId, err)
					continue
				}
				p := FSTreeResponse{*dir}
				fmt.Printf("%s\n", p.ToString(simplifyId))
			}
		}
	}else{
		fmt.Printf("error: you need specify a node id\n")
	}
}
func (c *console) FSCat(params map[string]interface{}) {
	ins, _ := params["inode"];
	if prm, ok := ins.([]string); ok && len(prm) > 0 {
		existIns := c.dashboard.GetInstances()
		collectMatch := map[string]string{}
		aliasMatch := []string{}
		for _, insIdAlias := range prm {
			// check to ins list
			for k, v := range existIns {
				if strings.Contains(k, insIdAlias) {
					aliasMatch   = append(aliasMatch, insIdAlias)
					collectMatch[k] = v.Name
				}
			}
		}
		filePath := os.Args[len(os.Args)-1]
		if len(filePath) == 0 {
			fmt.Printf("error: need file path")
			return
		}
		// is filePath equal one of the matched instance
		for _, iId := range aliasMatch {
			if iId == filePath {
				fmt.Printf("error: you need to specify file path!\n")
				return
			}
		}
		if len(collectMatch) == 0 {
			fmt.Printf("error: none instance selected\n")
		}
		for simplifyId, iId := range collectMatch {
			if raw, _ := params["raw"]; raw.(bool) {
				if len(collectMatch) > 1 {
					fmt.Printf("error: you cant do -raw, selected %d instance(s)\n", len(collectMatch))
					return
				}
				b, err := c.dashboard.FSCat(iId, filePath)
				if err != nil {
					fmt.Printf("%s: error: %s\n", simplifyId, err)
					return
				}
				os.Stdout.Write(b)
			}else if save, _ := params["save"]; save.(bool) {
				if len(collectMatch) > 1 {
					fmt.Printf("error: you cant do -save, selected %d instance(s)\n", len(collectMatch))
					return
				}
				b, err := c.dashboard.FSCat(iId, filePath)
				if err != nil {
					fmt.Printf("%s: error: %s\n", simplifyId, err)
					return
				}
				_, fileName := filepath.Split(filePath)
				f, err := os.Create(fileName)
				if err != nil {
					fmt.Printf("error: %s\n", err)
					return
				}
				defer f.Close()
				n, err := f.Write(b)
				fmt.Printf("%s: wrote %d bytes -> %s\n", simplifyId, n, fileName)
			}else{
				_, err := c.dashboard.FSCat(iId, filePath)
				if err != nil {
					fmt.Printf("%s: error: %s\n", simplifyId, err)
				}else{
					fmt.Printf("%s: success\n", simplifyId)
				}
			}
		}
	}else{
		fmt.Printf("error: you need specify a node id\n")
	}
}
func (c *console) FSPut(params map[string]interface{}) {
	ins, _ := params["inode"];
	if prm, ok := ins.([]string); ok && len(prm) > 0 {
		existIns := c.dashboard.GetInstances()
		collectMatch := map[string]string{}
		aliasMatch := []string{}
		for _, insIdAlias := range prm {
			// check to ins list
			for k, v := range existIns {
				if strings.Contains(k, insIdAlias) {
					aliasMatch   = append(aliasMatch, insIdAlias)
					collectMatch[k] = v.Name
				}
			}
		}
		filePath := os.Args[len(os.Args)-1]
		if len(filePath) == 0 {
			fmt.Printf("error: need file path")
			return
		}
		// is filePath equal one of the matched instance
		for _, iId := range aliasMatch {
			if iId == filePath {
				fmt.Printf("error: you need to specify file path!\n")
				return
			}
		}
		if len(collectMatch) == 0 {
			fmt.Printf("error: none instance selected\n")
			return
		}
		// inp src:dst 
		// ex: local.txt:/root
		// ex: /root/.ssh/authorized_key:/root/.ssh
		splitted := strings.Split(filePath, ":")
		if len(splitted) != 2 {
			fmt.Printf("error: wrong parameter src:dst format\n")
			return
		}
		srcPath := splitted[0]
		dstPath := filepath.Join(splitted[1], filepath.Base(srcPath))
		dstPath = strings.Replace(dstPath, "\\", "/", -1)
		dt, err := ioutil.ReadFile(srcPath)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return
		}
		for simplifyId, iId := range collectMatch {
			err := c.dashboard.FSPut(iId, dstPath, dt)
			if err != nil {
				fmt.Printf("%s: error: %s\n", simplifyId, err)
				continue
			}
			fmt.Printf("%s: success %s -> %s\n", simplifyId, srcPath, dstPath)
		}
	}else{
		fmt.Printf("error: you need specify a node id\n")
	}
}
func (c *console) Exec(params map[string]interface{}) {
	ins, _ := params["inode"];
	if prm, ok := ins.([]string); ok && len(prm) > 0 {
		existIns := c.dashboard.GetInstances()
		collectMatch := map[string]string{}
		aliasMatch := []string{}
		for _, insIdAlias := range prm {
			// check to ins list
			for k, v := range existIns {
				if strings.Contains(k, insIdAlias) {
					aliasMatch   = append(aliasMatch, insIdAlias)
					collectMatch[k] = v.Name
				}
			}
		}

		lastAliasMatch := aliasMatch[len(aliasMatch)-1]
		lastM := -1
		for i, arg := range os.Args {
			if arg == lastAliasMatch {
				lastM = i + 1
			}
		}
		if len(collectMatch) == 0 || lastM == -1 || lastM >= len(os.Args) {
			fmt.Printf("error: none instance selected\n")
			return
		}
		execCmds := os.Args[lastM:]
		if len(execCmds) == 0 {
			fmt.Printf("error: need cmds\n")
			return
		}
		for simplifyId, iId := range collectMatch {
			res, err := c.dashboard.Exec(iId, execCmds)
			if err != nil {
				fmt.Printf("%s: error: %s\n", simplifyId, err)
				continue
			}
			fmt.Printf("%s: success ExitCode=%d\n", simplifyId, res.ExitCode)
		}
	}else{
		fmt.Printf("error: you need specify a node id\n")
	}
}
func (c *console) CreateInstance(params map[string]interface{}) {
	i, err := c.dashboard.CreateInstanceDefault()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}
	fmt.Printf("%.20s  %.14s  %.8s  %-15s %-15s %-15s\n%s  %s\n", 
		i.Name, 
		i.Image,
		i.Hostname,
		i.IP,
		i.RoutableIP,
		i.SessionHost,
		i.SessionId,
		i.ProxyHost,
	)
}
func (c *console) ListInstance(params map[string]interface{}) {
	ins := c.dashboard.GetInstances()
	if ins == nil || len(ins) == 0 {
		fmt.Printf("error: %s\n", "no instance exists")
		return
	}
	for k,v := range ins {
		if quiet, _ := params["q"]; quiet.(bool) {
			fmt.Printf("%.20s\n", k)
		} else if adv, _ := params["adv"]; adv.(bool) {
			fmt.Printf("%.20s  %s\n", k, v.ProxyHost)
		}else{
			fmt.Printf("%.20s  %.14s  %.8s  %-15s %-15s %-15s\n", 
				k, v.Image,
				v.Hostname,
				v.IP,
				v.RoutableIP,
				v.SessionHost,
			)
		}
	}
	//c.Log("%+#v\n", ins)
}
func (c *console) RemoveInstance(params map[string]interface{}) {
	ins, _ := params["inode"];
	if prm, ok := ins.([]string); ok && len(prm) > 0 {
		existIns := c.dashboard.GetInstances()
		for _, insIdAlias := range prm {
			// check to ins list
			for k, v := range existIns {
				if strings.Contains(k, insIdAlias) {
					err := c.dashboard.DeleteInstance(v.Name)
					if err != nil {
						fmt.Printf("%s: error: %s\n", k, err)
					}else{
						fmt.Printf("%s: success\n", k)
					}
				}
			}
		}
	}else{
		fmt.Printf("error: you need specify a node id\n")
	}
}
func (c *console) InteractiveCLI(params map[string]interface{}) {
	c.Log("Doing interactive cli...\nimplement soon\n")
}
func (c *console) WatchWS(params map[string]interface{}) {
	c.Log("Doing wss:// ...\n")
	c.Log("err: %s\n", c.dashboard.FetchOnWebSocket())
}
func (c *console) Version(params map[string]interface{}) {
	fmt.Println("v1.0")
}
func (c *console) About(params map[string]interface{}) {
	fmt.Println("Author CLI\n  > https://github.com/ii64\n(c) 2020")
}
func (c *console) Help(params map[string]interface{}) {
	c.Log("< PWD Panel Host:  %s\n", c.dashboard.system.Domain)
	c.Log("< PWD Session:     %s\n", c.dashboard.session.ID)
	c.Log("< PWD CreateAt:    %s\n", c.dashboard.session.CreateAt)
	c.Log("< PWD ExpiresAt:   %s\n", c.dashboard.session.ExpiresAt)
	c.Log("< PWD Host:        %s\n", c.dashboard.session.Host)
	c.Log("< PWD GWIPAddr:    %s\n", c.dashboard.session.PWDIPAddress)
	c.Log("< PWD Cur Inst:    %d\n", len(c.dashboard.GetInstances()))
	m := `Help:
    -h create          create new instance
    -h nodes           show instance list
    -i  -q              * id only
    -i  -adv            * proxy only
    -h rm              remove instances
    -i  <ins>           * instance names
    -h fstree          show fs tree
    -i  <ins>           * instances name
    -i  <dirPath>       * optional dir path
    -h fscat           show content file
    -i  -raw            * direct print stdout
    -i  -save           * save to local
    -i  <ins>           * instances name
    -i  <filePath>      * file path
    -h fsput           put file to instances
    -i  <ins>           * instances name
    -i  <src:dst>       * source to dst path
    -h exec            execute command
    -i  <ins>           * instances name
    -i  <cmd>           * command to execute
    -h watch           watch instances status
    -h help            show help
    -h about           show about
    -h version         show version
@ii64
`	
	binName := os.Args[0]
	m = strings.Replace(m, "-h", binName, -1)
	m = strings.Replace(m, "-i", strings.Repeat(" ", len(binName)), -1)
	fmt.Printf(m)
}
func (c *console) Serve() {
	flag.StringVar(&c.pwd_url, "pwd-url", os.Getenv("PWD_URL"), "PWD session url")

//	flag.StringVar()

	flag.Parse()
	if len(c.pwd_url) == 0 {
		c.Log("error: need pwd session url\n")
		return
	}
	d, err := NewDashboard(c.pwd_url)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}
	c.dashboard = d

	flg := map[string]func(map[string]interface{}){
		"create":      c.CreateInstance,
		"nodes":       c.ListInstance,
		"rm":          c.RemoveInstance,
		"exec":        c.Exec,
		"watch":       c.WatchWS,
		"fstree":      c.FSTreeList,
		"fscat":       c.FSCat,
		"fsput":       c.FSPut,
		"interactive": c.InteractiveCLI,
		"version":     c.Version,
		"about":       c.About,
		"help":        c.Help,
	}
	prm := map[string]interface{}{
		"q": false, // quiet
		"adv": false,
		"raw": false,
		"save": false,

		// other

	}
	params := []string{}
	for _, argv := range os.Args[1:] {
		if argv[0] == 45 && len(argv) > 0 {
			if _, exist := prm[argv[1:]]; exist {
				prm[argv[1:]] = true
			}
		}else{
			params = append(params, argv)
		}
	}
	if len(os.Args) > 1 { // if length more than 1
		for i, arg := range os.Args[1:] {
			if arg[0] != 45 && i == 0 { // if prefix is not `-`
				if c, exist := flg[arg]; exist {
					if len(params) > 1 {
						 prm["inode"] = params[1:]
					}else{
						prm["inode"] = []string{}
					}
					c(prm)
					return
				}
			}
		}
	}
	c.Help(prm)
}