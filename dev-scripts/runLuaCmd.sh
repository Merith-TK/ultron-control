
CMD1="[\""
echo "Please input your command:"
read CMDU
CMD2="\"]"

finalCMD=$CMD1$CMDU$CMD2

echo $finalCMD

if [[ $ultronURL == "" ]]; then
    echo "Please input Ultron API Address"
    read ultronURL
fi
if [[ $turtleID == "" ]]; then
    echo "Please input which turtle to send command to"
    read turtleID
fi

curl --header "Content-Type: application/json" \
  --request POST \
  --data "$finalCMD" \
  $ultronURL/turtle/$turtleID