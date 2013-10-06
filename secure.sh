PROBLEM="1"
CODE="x = raw_input()
print x"
$string
COUNTER=1
for i in $(ls question/1/*.in)
do
	CASE="PASS"
	EXTRA=""
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer=$answer".ans"
	DIFF=$(diff <(cat $i | python run_python_secure.py "$CODE") <(cat $answer))
	if [ "$DIFF" != "" ] 
	then
	    CASE="FAIL\n"
	    EXTRA=$(cat $i)
	fi
	echo -e "Test Case $COUNTER: $CASE$EXTRA"
	COUNTER=$[$COUNTER +1]
done
