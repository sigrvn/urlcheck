# urlcheck

A tool to check the status of URLs on webpages

## Example
```bash
$ urlcheck sigrvn.github.io
2022/07/19 12:07:11 [urlcheck] | Fetching links for url 'sigrvn.github.io' ...
2022/07/19 12:07:11 [urlcheck] | No protocol specified for url 'sigrvn.github.io', assuming HTTPS
2022/07/19 12:07:11 [urlcheck] | OK 'https://fonts.googleapis.com/css?family=Raleway:400,300,600' (code: 200, response time: 100.878917ms)
2022/07/19 12:07:11 [urlcheck] | OK 'https://sigrvn.github.io/css/normalize.css' (code: 200, response time: 23.161086ms)
2022/07/19 12:07:11 [urlcheck] | OK 'https://sigrvn.github.io/css/skeleton.css' (code: 200, response time: 21.346699ms)
2022/07/19 12:07:11 [urlcheck] | OK 'https://sigrvn.github.io/images/favicon.png' (code: 200, response time: 20.509245ms)
2022/07/19 12:07:11 [urlcheck] | OK 'https://github.com/sigrvn' (code: 200, response time: 694.780643ms)
2022/07/19 12:07:12 [urlcheck] | BROKEN 'https://www.linkedin.com/in/arveen-emdad/' (code: 999, response time: 476.507718ms)
2022/07/19 12:07:12 [urlcheck] | Finished checking urls for url 'https://sigrvn.github.io'.
2022/07/19 12:07:12 [urlcheck] |        Checked 6 urls, 5 OK, 1 BROKEN
```
