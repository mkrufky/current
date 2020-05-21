if [ `curl -s --header "Content-Type: application/json" --request POST --data '{"userId":"user1","name":"McDonalds"}' localhost:8080/visit` == '{"visitId":"some-visit-id-1"}' ]; then
echo pass
else
echo fail
fi
if [ `curl -s "localhost:8080/visit?visitId=some-visit-id-1"` == '[{"userId":"user1","name":"McDonalds","visitId":"some-visit-id-1"}]' ]; then
echo pass
else
echo fail
fi
if [ `curl -s --header "Content-Type: application/json" --request POST --data '{"userId":"user1","name":"Starbucks"}' localhost:8080/visit` == '{"visitId":"some-visit-id-2"}' ]; then
echo pass
else
echo fail
fi
if [ `curl -s "localhost:8080/visit?userId=user1&searchString=MCDONALDS_LAS_VEGAS"` == '[{"userId":"user1","name":"McDonalds","visitId":"some-visit-id-1"}]' ]; then
echo pass
else
echo fail
fi
if [ `curl -s --header "Content-Type: application/json" --request POST --data '{"userId":"user2","name":"Starbucks"}' localhost:8080/visit` == '{"visitId":"some-visit-id-3"}' ]; then
echo pass
else
echo fail
fi
if [ `curl -s "localhost:8080/visit?userId=user2&searchString=APPLE"` == '[]' ]; then
echo pass
else
echo fail
fi