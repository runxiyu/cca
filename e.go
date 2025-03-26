package main

import (
	"context"
	"strconv"
	"strings"
)

func eee(ctx context.Context) (res []student_ish, err error) {
	ni, err := queryNameID(ctx, "SELECT name, id FROM expected_students")
	if err != nil {
		return nil, wrapError(errUnexpectedDBError, err)
	}

	rows, err := db.Query(
		ctx,
		"SELECT name, email, department, confirmed FROM users",
	)
	if err != nil {
		return nil, wrapError(errUnexpectedDBError, err)
	}
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return nil, wrapError(errUnexpectedDBError, err)
			}
			break
		}
		var currentUserName, currentEmail, currentDepartment string
		var currentConfirmed bool
		err := rows.Scan(
			&currentUserName,
			&currentEmail,
			&currentDepartment,
			&currentConfirmed,
		)
		if err != nil {
			return nil, wrapError(errUnexpectedDBError, err)
		}
		unamepart, _, _ := strings.Cut(currentEmail, "@")
		unamepart = strings.TrimPrefix(strings.TrimPrefix(unamepart, "s"), "S")
		nii, _ := strconv.ParseInt(unamepart, 10, 64)
		delete(ni, nii)

		if currentDepartment == staffDepartment {
			continue
		}

		res = append(
			res,
			student_ish{
				Name:       currentUserName,
				Email:      currentEmail,
				Department: currentDepartment,
				Status:     strconv.FormatBool(currentConfirmed),
			},
		)
	}

	for k, v := range ni {
		/*
			res = append(
				res,
				[]string{
					v,
					"s" + strconv.FormatInt(k, 10) + "@ykpaoschool.cn",
					"Unknown",
					"never logged in",
				},
			)*/
		res = append(
			res,
			student_ish{
				Name:       v,
				Email:      "s" + strconv.FormatInt(k, 10) + "@ykpaoschool.cn",
				Department: "Unknown",
				Status:     "never logged in",
			},
		)
	}

	return
}
