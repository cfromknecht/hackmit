PROBLEM="1"
CODE="x = raw_input()
print x"

COUNTER=1
for i in $(ls question/1/*.in)
do
	CASE="PASS"
	EXTRA=""
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer=$answer".ans"
	output=$(cat $i | python run_python_secure.py "$CODE")
	DIFF=$(diff <(echo $output) <(cat $answer))
	if [ "$DIFF" != "" ] 
	then
	    CASE="FAIL\n"
	    EXTRA="returned\n$output\ninput\n"$(cat $i)
	    # EXTRA+=$(cat $i)
	fi
	echo -e "Test Case $COUNTER: $CASE$EXTRA\n\n"
	COUNTER=$[$COUNTER +1]
done
