PROBLEM="$1"
CODE="$2"

COUNTER=1
for i in $(ls question/1/*.in)
do
	STATUS="PASS"
	EXTRA=""
	answer=$(echo $i | rev | cut -c 4- | rev)
	answer=$answer".ans"
	OUTPUT=$(cat $i | python run_python_secure.py "$CODE" | perl -pe 's/\\/\\\\/g' | perl -pe 's/"/\\"/g')
	INPUT=$(cat $i)
	DIFF=$(diff <(echo $output) <(cat $answer))
	if [ "$DIFF" != "" ] 
	then
	    STATUS="FAIL"
	fi
	printf "{\"status\":\"$STATUS\", \"output\":'$OUTPUT', \"input\":'$INPUT'}"
done
