Date = 2020-09-11T13:37:00-04:00
Live = false
[Meta]
Title = "Takeaways from Quantitative Trading by Earnest P. Chan"
Description = "I read a book on quantitative finance and took some notes..."
---
## Introduction

I read a book! It's actually super rare for me to read a book. I read a bit online everyday, stuff like articles, blogs, and code documentation. I really enjoy writing more than reading so I end up writing a lot of code. Reading is usually just a means to more writing as it was here. 

I read Quantitative Trading by Earnest P. Chan. I do a lot of programming and want to learn more about finance, so it seems like a good intersection of the two. I also hear quantitative traders can make bank working at a hedgefund, so if I have a 7 figure salary in a few years, I won't complain. 

This post is to lay out my notes for important takeaways from the book, and also for me to explain some of the concepts to make sure I really understand them *cough* sharpe ratios *cough*. The book has been super approachable and a pleasure to read. After I lay out the concepts here, the next post will be a backtesting and trading system design, then I'll get to start implementing it. I'm interested in equities to start since that is what I understand the best, but long term I'd like to use this is a platform to understand derivatives, fixed income, and forex. If it goes decently in real money, I might also take a look at magical internet money.

## Concepts

### Sharpe Ratio
* TODO: Understand what is the risk free rate in sharpe
* Annualizing a stddev * sqrt(12) for sharpe
* Using Sharpe
 - Annual sharpe
 - Gt 3 daily profitable
 - Gt 2 monthly profitable
 - Sub 1 not viable

### Backtesting

* Data sources

* Backtest biases
 - Survivorship
 - Spread changes
 - Splits
 - Dividends

- Measure drawdowns length and depth
- Daily returns outside of 4 stddev are sus

### Sensitivity analysis

- Modify parameters
- If it falls apart it was shit
- Eliminate parameters

### Allocate capital over parameter variations

### Strategies

- Pairs trading
- Mean reversion
- Momentum

### Optimal Capital Allocation

* Kelly formula

### Misc notes

* Spread Trader
* Don't trade low priced stocks
 - Fee per share, higher bid ask spread
* Don't exceed 1% of avg daily vol
* Scale to market cap
* Batch orders to avoid slippage
* Paper trade
* 
* Beautiful Soup in Golang

