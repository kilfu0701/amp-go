package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"github.com/yosssi/gohtml"
)

func main() {
	htmlString := `
<html>
  <head>
    <meta charset="utf-8">
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.0/jquery.min.js"></script>
  </head>
  <body>
    <h1>Test</h1>
    <p>Contents here</p>
    <img src="http://i2.cdn.turner.com/cnnnext/dam/assets/160208081229-gaga-superbowl-exlarge-169.jpg">
    <iframe src="https://www.reddit.com"></iframe>
  </body>
</html>`

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		log.Fatal(err)
	}

	styleAmpBoilerplateNode := &html.Node{
		FirstChild: &html.Node{
			Type:     html.TextNode,
			Data:     "body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-moz-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-ms-animation:-amp-start 8s steps(1,end) 0s 1 normal both;animation:-amp-start 8s steps(1,end) 0s 1 normal both}@-webkit-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-moz-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-ms-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-o-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}",
			DataAtom: atom.Body,
		},

		Type:     html.ElementNode,
		Data:     "style",
		DataAtom: atom.Style,
		Attr:     []html.Attribute{
			{Key: "amp-boilerplate"},
		},
	}

	noscriptNode := &html.Node{
		FirstChild: &html.Node{
			FirstChild: &html.Node{
				Type:     html.TextNode,
				Data:     "body{-webkit-animation:none;-moz-animation:none;-ms-animation:none;animation:none}",
				DataAtom: atom.Body,
			},

			Type:     html.ElementNode,
			Data:     "style",
			DataAtom: atom.Style,
			Attr:     []html.Attribute{
				{Key: "amp-boilerplate"},
			},
		},

		Type:     html.ElementNode,
		Data:     "noscript",
		DataAtom: atom.Noscript,
	}

	needAddMetaCharset := true
	needAddDoctype := true
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch(n.Data) {
				// add amp into html
				case "html":
					n.Attr = []html.Attribute{
						{Key: "amp"},
					}

				// add needed tags into head
				case "head":
					n.AppendChild(styleAmpBoilerplateNode)
					n.AppendChild(noscriptNode)

				case "img", "iframe", "video", "audio":
					n.Data = "amp-" + n.Data

				case "meta":
					for _, attr := range n.Attr {
						if attr.Key == "charset" {
							needAddMetaCharset = false
						}
					}
			}

			// remove all script tags
			if n.Data == "script" {
				n.Parent.RemoveChild(n)
			}
		} else if n.Type == html.DoctypeNode {
			if n.Data == "html" {
				needAddDoctype = false
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if needAddMetaCharset {
		metaCharsetNode := &html.Node{
			Type:     html.ElementNode,
			Data:     "meta",
			DataAtom: atom.Meta,
			Attr:     []html.Attribute{
				{Key: "charset", Val: "utf-8"},
			},
		}
		var fn func(*html.Node)
		fn = func(n *html.Node) {
			if n.Type == html.ElementNode {
				switch(n.Data) {
					case "head":
						n.AppendChild(metaCharsetNode)
						return
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				fn(c)
			}
		}
		fn(doc)
	}

	buf := bytes.NewBufferString("")
	if err := html.Render(buf, doc); err != nil {
		log.Fatal(err)
	}

	result := buf.String()
	if needAddDoctype {
		result = "<!DOCTYPE html>" + result
	}

	fmt.Println(gohtml.Format(result))
}
