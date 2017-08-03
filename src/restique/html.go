package main

const PAGE = `
<html>
<head>
<title>RESTIQUE {{VERSION}}</title>
<style>
.form {
  position: relative;
  z-index: 1;
  background: #FFFFFF;
  max-width: 360px;
  margin: 0 auto 100px;
  padding: 45px;
  padding-bottom:30px;
  text-align: center;
  box-shadow: 0 0 20px 0 rgba(0, 0, 0, 0.2), 0 5px 5px 0 rgba(0, 0, 0, 0.24);
}
.form input {
  outline: 0;
  background: #f2f2f2;
  width: 100%;
  border: 0;
  margin: 0 0 15px;
  padding: 15px;
  box-sizing: border-box;
  font-size: 14px;
}
.form button {
  text-transform: uppercase;
  outline: 0;
  background: #4CAF50;
  width: 100%;
  border: 0;
  padding: 15px;
  color: #FFFFFF;
  font-size: 14px;
  -webkit-transition: all 0.3 ease;
  transition: all 0.3 ease;
  cursor: pointer;
}
.form button:hover,.form button:active,.form button:focus {
  background: #43A047;
}
.headrow {background:#666666;color:white}
.evenrow {background:#f8f8f8}
.oddrow {background:#e8e8e8}
.thcell {padding:6px;text-align:left}
.tdcell {padding:6px;vertical-align:top}
.oddhist {padding:5px;background:#e8e8e8;cursor:pointer}
.evenhist {padding:5px;background:#f8f8f8;cursor:pointer}
.oddhist:hover {background:white}
.evenhist:hover {background:white}
</style>
</head>
<body>
{{CONTENT}}
</body>
</html>
`
