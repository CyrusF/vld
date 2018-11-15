#!/bin/bash
declare -i n
n=0
function walk_dir()
{
	for file in `ls $2`
	do
		if [ -d "$2/$file" ]; then
			walk_dir $1 "$2/$file"
		else
			if [ "${file##*.}" = "php" ]; then
				echo "Processing "$2"/"$file
				n=$n+1
				`php -dvld.active=1 -dvld.execute=$1 -dvld.webshell_test $2/$file 1>/dev/null 2>>log`
			else
				echo "Skip       "$2"/"$file
			fi
		fi
	done
}   
rm log
START=$(date "+%s");
echo "Execute = $1"
echo "Target  = $2"
walk_dir $2 $1
END=$(date "+%s");
time=$((END-START))
echo "Total webshell:"
cat log | grep "WEBSHELL" | wc -l
echo "Done. Test "$n" files in "$time" s."

# rm log
