# urlcheck

A CLI tool to check the status of URLs on webpages

## Example output
```bash
$ urlcheck sigrvn.github.io
2022/07/19 17:20:25 [urlcheck] | Using 1 worker(s)
2022/07/19 17:20:25 [urlcheck] | Fetching links for url 'sigrvn.github.io'
2022/07/19 17:20:25 [urlcheck] | No protocol specified for url 'sigrvn.github.io', assuming HTTPS
2022/07/19 17:20:25 [urlcheck] | OK(200) 'https://fonts.googleapis.com/css?family=Raleway:400,300,600' (response time: 98.541623ms)
2022/07/19 17:20:25 [urlcheck] | OK(200) 'https://sigrvn.github.io/css/normalize.css' (response time: 23.542996ms)
2022/07/19 17:20:25 [urlcheck] | OK(200) 'https://sigrvn.github.io/css/skeleton.css' (response time: 21.294966ms)
2022/07/19 17:20:25 [urlcheck] | OK(200) 'https://sigrvn.github.io/images/favicon.png' (response time: 26.192488ms)
2022/07/19 17:20:25 [urlcheck] | OK(200) 'https://github.com/sigrvn' (response time: 305.759456ms)
2022/07/19 17:20:26 [urlcheck] | BROKEN(999) 'https://www.linkedin.com/in/arveen-emdad/' (response time: 434.756577ms)
2022/07/19 17:20:26 [urlcheck] | Finished checking urls for 'https://sigrvn.github.io' in 910.088106ms.
2022/07/19 17:20:26 [urlcheck] |        Checked 6 urls, 5 OK, 1 BROKEN
```

## Notes

* Parses links from anchor tags only, inlined content is not parsed
* Does not work with Single Page Applications as JavaScript is not executed
