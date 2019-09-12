-- REPEAT 10
-- NAME owner
insert into "owner" ("name", "date_of_birth") values
{{range $i, $e := $.times_10 }}
	{{if $i}},{{end}}
	(
		'{{stringf "%s.%s@acme.co.uk" 5 5 "abcdefg" 5 5 "hijklmnop" }}',
		'{{date "1900-01-01" "now" ""}}'
	)
{{end}}
returning "id", "name", "date_of_birth";

-- REPEAT 20
-- NAME pet
insert into "pet" ("pid", "name", "owner_name", "owner_date_of_birth") values
{{range $i, $e := .times_10 }}
	{{if $i}},{{end}}
	(
		'{{each "owner" "id" $i}}',
		'{{string 10 10 "p-" "abcde"}}',
		'{{each "owner" "name" $i}}',
		'{{each "owner" "date_of_birth" $i}}'
	)
{{end}};

-- EOF

