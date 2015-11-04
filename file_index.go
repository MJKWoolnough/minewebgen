package main

import (
	"time"

	"github.com/MJKWoolnough/httpdir"
)

func init() {
	dir.Create("index.html", httpdir.FileString(`<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<title>Minecraft Generator</title>
		<script type="text/javascript" src="js/js.js"></script>
	</head>
	<body>Connecting...</body>
</html>
`, time.Unix(1446640983, 0)))
}
