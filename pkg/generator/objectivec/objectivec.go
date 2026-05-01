package objectivec

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var code string
	code += "#import <Foundation/Foundation.h>\n"
	code += "\n"
	code += fmt.Sprintf("NSURL *url = [NSURL URLWithString:%s];\n", reprStr(r.URLs[0].URL))

	if r.BasicAuth != "" {
		code += "\n"
		user, pass := splitBasicAuth(r.BasicAuth)
		code += fmt.Sprintf("NSString *credentials = [NSString stringWithFormat:@\"%%@:%%@\", %s, %s];\n", reprStr(user), reprStr(pass))
		code += "NSData *credentialsData = [credentials dataUsingEncoding:NSUTF8StringEncoding];\n"
		code += "NSString *base64Credentials = [credentialsData base64EncodedStringWithOptions:0];\n"
	}

	reservedHeaders := []string{
		"content-length", "authorization", "connection", "host",
		"proxy-authenticate", "proxy-authorization", "www-authenticate",
	}

	if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
		headerLines := []string{}
		for _, h := range r.HeaderKV {
			if h.Value == "" {
				continue
			}
			lowerKey := strings.ToLower(h.Key)
			for _, reserved := range reservedHeaders {
				if lowerKey == reserved {
					// Warning would be added here in a full implementation
					break
				}
			}
			headerLines = append(headerLines, fmt.Sprintf("    %s: %s", reprStr(h.Key), reprStr(h.Value)))
		}
		if r.BasicAuth != "" {
			headerLines = append(headerLines, `    @"Authorization": [NSString stringWithFormat:@"Basic %@", base64Credentials]`)
		}
		if len(headerLines) > 0 {
			code += "NSDictionary *headers = @{\n"
			code += strings.Join(headerLines, ",\n") + "\n"
			code += "};\n"
			code += "\n"
		}
	}

	hasData := false
	hasDataStream := false
	if r.BodyFile != "" {
		code += fmt.Sprintf("InputStream *dataStream = [InputStream fileAtPath:%s];", reprStr(r.BodyFile))
		hasDataStream = true
	} else if len(r.FormParts) > 0 {
		// TODO: multipart uploads
	} else if r.Body != "" {
		parts := strings.Split(r.Body, "&")
		if len(parts) > 1 {
			encodedParts := []string{}
			for i, part := range parts {
				if i > 0 {
					part = "&" + part
				}
				encodedParts = append(encodedParts, reprStr(part))
			}
			code += fmt.Sprintf("NSMutableData *data = [[NSMutableData alloc] initWithData:[%s dataUsingEncoding:NSUTF8StringEncoding]];\n", encodedParts[0])
			for i := 1; i < len(encodedParts); i++ {
				code += fmt.Sprintf("[data appendData:[%s dataUsingEncoding:NSUTF8StringEncoding]];\n", encodedParts[i])
			}
			code += "\n"
			hasData = true
		} else {
			code += fmt.Sprintf("NSString *data = @\"%s\";\n", r.Body)
			hasData = true
		}
	}

	if r.MaxTime != "" && r.MaxTime != "60" && r.MaxTime != "60.0" {
		code += "NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:url\n"
		code += "                                                          cachePolicy:NSURLRequestUseProtocolCachePolicy\n"
		code += fmt.Sprintf("                                                      timeoutInterval:%s];\n", r.MaxTime)
	} else {
		code += "NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:url];\n"
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	if strings.ToUpper(method) != "GET" {
		code += fmt.Sprintf("[request setHTTPMethod:%s];\n", reprStr(method))
	}
	if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
		code += "[request setAllHTTPHeaderFields:headers];\n"
	}
	if hasData {
		code += "[request setHTTPBody:data];\n"
	} else if hasDataStream {
		code += "[request setHTTPBodyStream:dataStream];\n"
	}

	code += "\n"
	code += "NSURLSessionConfiguration *defaultSessionConfiguration = [NSURLSessionConfiguration defaultSessionConfiguration];\n"
	code += "NSURLSession *session = [NSURLSession sessionWithConfiguration:defaultSessionConfiguration];\n"
	code += "NSURLSessionDataTask *dataTask = [session dataTaskWithRequest:request completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {\n"
	code += "    if (error) {\n"
	code += "        NSLog(@\"%@\", error);\n"
	code += "    } else {\n"
	code += "        NSHTTPURLResponse *httpResponse = (NSHTTPURLResponse *) response;\n"
	code += "        NSLog(@\"%@\", httpResponse);\n"
	code += "    }\n"
	code += "}];\n"
	code += "[dataTask resume];\n"

	return code
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\a":
			return "\\a"
		case "\b":
			return "\\b"
		case "\f":
			return "\\f"
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\t":
			return "\\t"
		case "\v":
			return "\\v"
		case "\\":
			return "\\\\"
		case "\"":
			return "\\\""
		default:
			if len(c) == 1 {
				hex := fmt.Sprintf("%02X", c[0])
				return "\\x" + hex
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\u" + hex
		}
	})
	return "@\"" + escaped + "\""
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
