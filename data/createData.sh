
#data format date-service-value
function main() {
	echo "main start"
	for (( i = 1; i <= 100; i++ ))
	do
		echo "cnt"
		data $1 $((i)) $3
	done
	echo "main end"
}

function data() {
	filename=data/$1log$2.log
	cnt=$3
	echo "$1 $2 $3"
	for (( i = 1; i <= ${cnt}; i++ ))
	do
		echo "$(date)-${filename}-$((RANDOM%100)).$((RANDOM%100))" >> ${filename}
	done
}


#name nums cnt
data $1 $2 $3
