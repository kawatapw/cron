package main

import (
	"fmt"

	"github.com/fatih/color"
)

func fixStatsOverflow(relax bool) {
	defer wg.Done()
	var table string
	if relax {
		table = "users_stats_relax"
	} else {
		table = "users_stats"
	}
	initQuery := fmt.Sprintf(`SELECT id FROM %s JOIN users USING(id) WHERE 
	ranked_score_std < 0 OR 
	ranked_score_taiko < 0 OR 
	ranked_score_ctb < 0 OR 
	ranked_score_mania < 0`, table)
	rows, err := db.Query(initQuery)
	if err != nil {
		queryError(err, initQuery)
		return
	}
	var scoresCount int
	var usersCount int
	rankedScores := make(map[int][]int)
	for rows.Next() {
		var uid int
		err = rows.Scan(&uid)
		if err != nil {
			queryError(err, initQuery)
			continue
		}
		if usersCount%1000 == 0 {
			verboseln("> FixStatsOverflow::users:", usersCount)
		}
		const fetchQuery = "SELECT score, play_mode FROM scores JOIN beatmaps USING(beatmap_md5) WHERE userid = ? AND completed = 3 AND is_relax = ?"
		var relaxV int
		if relax {
			relaxV = 1
		}
		scoreRows, err := db.Query(fetchQuery, uid, relaxV)
		if err != nil {
			queryError(err, fetchQuery, uid)
			continue
		}
		for scoreRows.Next() {
			if scoresCount%1000 == 0 {
				verboseln("> FixStatsOverflow::scores:", scoresCount)
			}

			var score int
			var mode int
			scoreRows.Scan(&score, &mode)
			if mode < 0 || mode > 3 {
				continue
			}
			// fmt.Println(rankedScores)
			if rankedScores[uid] == nil {
				rankedScores[uid] = make([]int, 4)
			}
			rankedScores[uid][mode] += score
			scoresCount++
			if rankedScores[uid][mode] < 0 {
				verboseln("> FixStatsOverflow: overflow for user", uid, "(hax)! Breaking out of the loop.")
				rankedScores[uid] = nil
				break
			}
		}
		verboseln("> FixStatsOverflow: done", uid)
		usersCount++
	}
	for uid, v := range rankedScores {
		if v == nil {
			v = make([]int, 4)
		}
		op(fmt.Sprintf("UPDATE %s SET ranked_score_std = ?, ranked_score_taiko = ?, ranked_score_ctb = ?, ranked_score_mania = ? WHERE id = ? LIMIT 1", table), v[0], v[1], v[2], v[3], uid)
	}
	color.Green("> FixStatsOverflow: done!")
}

func opFixStatsOverflow() {
	fixStatsOverflow(false)
	fixStatsOverflow(true)
}
