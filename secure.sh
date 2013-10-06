PROBLEM="1"
CODE="import os
x = raw_input()
print x"

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
		echo $DIFF
	    CASE="FAIL\n"
	    EXTRA=$(cat $i)
	fi
	echo -e "Test Case $COUNTER: $CASE$EXTRA\n\n"
	COUNTER=$[$COUNTER +1]
done
