<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI-Powered Technical Analysis Dashboard</title>
    <style>
       body {
           font-family: sans-serif;
       }
       table {
           border-collapse: collapse;
           width: 50%;
       }
       th, td {
           border: 1px solid black;
           padding: 8px;
           text-align: left;
       }
    </style>
</head>
<body>
    <h1>AI-Powered Technical Analysis Dashboard</h1>

    <form action="/analyze" method="post">
        <label for="tickers">Stock Tickers (comma-separated):</label>
        <input type="text" id="tickers" name="tickers" value="{{.Config.Tickers | join ", "}}"><br><br>

        <label for="start_date">Start Date:</label>
        <input type="date" id="start_date" name="start_date" value="{{.Config.StartDate.Format "2006-01-02"}}"><br><br>

        <label for="end_date">End Date:</label>
        <input type="date" id="end_date" name="end_date" value="{{.Config.EndDate.Format "2006-01-02"}}"><br><br>

        <label>Technical Indicators:</label><br>
        <input type="checkbox" id="sma" name="indicators" value="20-Day SMA" {{if contains .Config.Indicators "20-Day SMA"}}checked{{end}}>
        <label for="sma">20-Day SMA</label><br>

        <input type="checkbox" id="ema" name="indicators" value="20-Day EMA" {{if contains .Config.Indicators "20-Day EMA"}}checked{{end}}>
        <label for="ema">20-Day EMA</label><br>

        <input type="checkbox" id="bb" name="indicators" value="20-Day Bollinger Bands" {{if contains .Config.Indicators "20-Day Bollinger Bands"}}checked{{end}}>
        <label for="bb">20-Day Bollinger Bands</label><br>

        <input type="checkbox" id="vwap" name="indicators" value="VWAP" {{ if contains .Config.Indicators "VWAP"}}checked{{end}}>
        <label for="vwap">VWAP</label><br><br>

        <input type="submit" value="Analyze">
    </form>

    <h2>Overall Summary</h2>
    <table>
        <thead>
            <tr>
                <th>Stock</th>
                <th>Recommendation</th>
            </tr>
        </thead>
        <tbody>
            {{range .TickerData}}
            <tr>
                <td>{{.Ticker}}</td>
                <td>{{.Result.Action}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <h2>Detailed Analysis</h2>
     {{range .TickerData}}
        <h3>Analysis for {{.Ticker}}</h3>
        <img src="data:image/png;base64,{{.Chart}}" alt="Chart for {{.Ticker}}"><br>
        <strong>Detailed Justification:</strong>
        <p>{{.Result.Justification}}</p>
    {{end}}

<script>
// Function to set the default values of the date pickers on page load
window.onload = function() {
    // Set Start Date to one year before today
    var today = new Date();
    var lastYear = new Date(today);
    lastYear.setFullYear(today.getFullYear() - 1);
    document.getElementById('start_date').value = lastYear.toISOString().split('T')[0];
    // Set End Date to today
    document.getElementById('end_date').value = today.toISOString().split('T')[0];
}
</script>
</body>
</html>
