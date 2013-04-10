package util

import (
	log "github.com/cihub/seelog"
	"strings"
)

import (
	"trygo/ssss"
)

func RenderError(c *ssss.Controller, err interface{}) {
	fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		RenderJsonError(c, err)
	case "xml":
		RenderXmlError(c, err)
	default:
		RenderJsonError(c, err)
	}
}

func RenderSucceed(c *ssss.Controller, data interface{}) {
	fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		RenderJsonSucceed(c, data)
	case "xml":
		RenderXmlSucceed(c, data)
	default:
		RenderJsonSucceed(c, data)
	}
}

func RenderJsonError(c *ssss.Controller, err interface{}) {
	rs := ConvertErrorResult(err)
	if rs.Code == ERROR_CODE_RUNTIME {
		log.Error(err)
	} else {
		log.Debug(err)
	}

	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		c.RenderJQueryCallback(jsoncallback, rs)
	} else {
		c.RenderJson(rs)
	}
}

func RenderJsonSucceed(c *ssss.Controller, data interface{}) {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		c.RenderJQueryCallback(jsoncallback, SucceedResult(data))
	} else {
		c.RenderJson(SucceedResult(data))
	}
}

func RenderXmlError(c *ssss.Controller, err interface{}) {
	rs := ConvertErrorResult(err)
	if rs.Code == ERROR_CODE_RUNTIME {
		log.Error(err)
	} else {
		log.Debug(err)
	}
	c.RenderXml(rs)
}

func RenderXmlSucceed(c *ssss.Controller, data interface{}) {
	c.RenderXml(SucceedResult(data))
}
