--- main lucky bet
-- bet your block and get reward!
-- @gas_limit 100000000
-- @gas_price 0
-- @param_cnt 0
-- @return_cnt 0
-- @publisher walleta
function main()
	Assert(Put("max_user_number", 100))
	Assert(Put("user_number", 0))
	Assert(Put("total_coins", 0))
	Assert(Put("last_lucky_block", -1))
	Assert(Put("round", 0))
	Assert(clearUserValue() == 0)
end--f

--- clearUserValue clear user bet value 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege private
function clearUserValue()
	clearTable = {}
	for i = 0, 9, 1 do
		userTableKey = string.format("user_value%d", i)
		ok, json = ToJson(clearTable)
		Assert(ok)
		Assert(Put(userTableKey, json))
	end
	return 0
end--f

--- Bet a lucky number
-- bet a lucky number with 1 ~ 5 coins
-- @param_cnt 4
-- @return_cnt 1
-- @privilege public
function Bet(account, luckyNumber, coins, nonce)
	if (not (coins >= 1 and coins <= 5))
	then
	    return "bet coins should be >=1 and <= 5"
	end
	if (not (luckyNumber >= 0 and luckyNumber <= 9))
	then
	    return "bet lucky number should be >=0 and <= 9"
	end
	Assert(nonce ~= nil)

	ok, maxUserNumber = Get("max_user_number")
	Assert(ok)
	ok, number = Get("user_number")
	Assert(ok)
	ok, totalCoins = Get("total_coins")
	Assert(ok)

	Assert(Deposit(account, coins) == true)
	userTableKey = string.format("user_value%d", luckyNumber)

	ok, json = Get(userTableKey)
	Assert(ok)
	ok, valTable = ParseJson(json)
	Assert(ok)
	if (valTable == nil)
	then
		valTable = {}
	end
	Assert(valTable ~= nil)

	len = #valTable
	valTable[len + 1] = account
	valTable[len + 2] = coins 
	valTable[len + 3] = nonce 

	ok, json = ToJson(valTable)
	Assert(ok)
	Assert(Put(userTableKey, json))

	number = number + 1
	totalCoins = totalCoins + coins
	Assert(Put("user_number", number))
	Assert(Put("total_coins", totalCoins))

	if (number >= maxUserNumber)
	then
		blockNumber = Height()
		ok, lastLuckyBlock = Get("last_lucky_block")
		Assert(ok)
		pHash = ParentHash()

		if (lastLuckyBlock < 0 or blockNumber - lastLuckyBlock >= 16 or blockNumber > lastLuckyBlock and pHash % 16 == 0)
		then
			Assert(Put("user_number", 0))
			Assert(Put("total_coins", 0))
			Assert(Put("last_lucky_block", blockNumber))
			Assert(getReward(blockNumber, totalCoins, number) == 0)
		end
	end

	return 0
end--f

--- getReward give reward to win users
-- @param_cnt 3
-- @return_cnt 1
-- @privilege private
function getReward(blockNumber, totalCoins, userNumber)
	luckyNumber = blockNumber % 10
	ok, round = Get("round")
	Assert(ok)
	round = round + 1
	roundKey = string.format("round%d", round)
	roundValue = ""

	userTableKey = string.format("user_value%d", luckyNumber)
	ok, json = Get(userTableKey)
	Assert(ok)
	ok, valTable = ParseJson(json)
	Assert(ok)
	if (valTable == nil)
	then
		valTable = {}
	end

	res = {}
	res["UnwinUserList"] = {}
	for i = 0, 9, 1 do
		if (i ~= luckyNumber)
		then
			userTableKey = string.format("user_value%d", i)
			ok, json = Get(userTableKey)
			Assert(ok)
			ok, tmpTable = ParseJson(json)
			Assert(ok)

			tn = #tmpTable
			kn = math.floor((tn + 1) / 3)
			for j = 0, kn - 1, 1 do
				a0 = tmpTable[j * 3 + 1]
				a1 = tmpTable[j * 3 + 2]
				a2 = tmpTable[j * 3 + 3]
				unwinUser = {}
				unwinUser["Address"] = a0
				unwinUser["BetAmount"] = a1
				unwinUser["Amount"] = 0
				unwinUser["Nonce"] = a2
				unwinUser["LuckyNumber"] = i
				l0 = #(res["UnwinUserList"])
				res["UnwinUserList"][l0 + 1] = unwinUser
			end
		end
	end

	Assert(clearUserValue() == 0)

	totalCoins = totalCoins * 0.95
	totalVal = 0
	len = #valTable
	kNumber = math.floor((len + 1) / 3)

	for i = 0, kNumber - 1, 1 do
		totalVal = totalVal + valTable[i * 3 + 2]
	end

	res["BlockHeight"] = blockNumber
	res["TotalUserNumber"] = userNumber
	res["WinUserNumber"] = kNumber
	res["TotalRewards"] = totalCoins
	res["WinUserList"] = {}

	for i = 0, kNumber - 1, 1 do
		address = valTable[i * 3 + 1]
		betCoins = valTable[i * 3 + 2]
		winUser = {}
		winUser["BetAmount"] = betCoins
		winUser["LuckyNumber"] = luckyNumber
		winCoins = betCoins * 1.0 / totalVal * totalCoins
		winUser["Address"] = address
		winUser["Amount"] = winCoins
		winUser["Nonce"] = valTable[i * 3 + 3]

		Assert(Withdraw(address, winCoins) == true)
		len = #(res["WinUserList"])
		res["WinUserList"][len + 1] = winUser
	end

	ok, roundValue = ToJson(res)
	Assert(ok)
	Assert(Put(roundKey, roundValue))
	Assert(Put("round", round))

	return 0
end--f

--- QueryUserNumber query user number now 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryUserNumber()
	ok, r = Get("user_number")
	Assert(ok)
	return r
end--f

--- QueryTotalCoins query total coins
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryTotalCoins()
	ok, r = Get("total_coins")
	Assert(ok)
	return r
end--f

--- QueryLastLuckyBlock query last lucky block 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryLastLuckyBlock()
	ok, r = Get("last_lucky_block")
	Assert(ok)
	return r
end--f

--- QueryMaxUserNumber query max user number 
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryMaxUserNumber()
	ok, r = Get("max_user_number")
	Assert(ok)
	return r
end--f

--- QueryRound query round
-- @param_cnt 0
-- @return_cnt 1
-- @privilege public
function QueryRound()
	ok, r = Get("round")
	Assert(ok)
	return r
end--f

