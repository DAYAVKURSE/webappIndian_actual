package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

var badPatchs = []string{
	".env", "DIAGNOSTICS", "ports", "console",
	"php", "login", "mysql", "admin", "cgi-bin", "index.jsp",
	"download", "powershell", "favicon.ico", "format=json", "actuator",
	"geoserver", "goform", "luci", "set_LimitClient_cfg", "manager", "wp-login.php",
	"wp-admin", "xmlrpc.php", "config.php", "passwd", "shadow", "backup", "secret",
	"usernames", "passwords", "confidential", "private", "bin/bash", "bin/sh",
	"cmd.exe", "administrator", "shell", "exec", "command", "query", "select",
	"insert", "delete", "update", "drop", "alter", "union", "concat", "password",
	"ftp", "tftp", "smb", "rpcbind", "bconsole", "tomcat", "manager/html", "web-console", "login.do",
}

func BlockBadActorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path

		for _, path := range badPatchs {
			if strings.Contains(requestPath, path) {
				c.JSON(403, gin.H{"error": "Forbidden"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
