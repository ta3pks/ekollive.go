package controllers

import (
	"fmt"
	"strconv"
	"sync"

	"../lib"
	"../lib/db"
	"../lib/store/oddids"
	"../models"
	"../models/odd"
	"../models/oddType"
	"../models/oddfieldType"
	"github.com/sanity-io/litter"
)

var oddsLock sync.Mutex

func UpsertOdds(match models.Match) {
	for _, o := range match.Odds {
		//each odd
		od := &oddType.Oddtype{
			Subtype:      o.Subtype,
			Type:         o.Type,
			Typeid:       o.Typeid,
			Oddtypevalue: o.Freetext,
			Oddtypeid:    o.OddTypeId,
		}
		if od.Oddtypeid == nil && oddids.Set(od) == 0 {
			fmt.Println("\n\nan error occured inserting the odd")
			litter.Dump(od)
		}
		//insert oddFields
		func(od oddType.Oddtype, o models.Odd) {
			if *lib.LockOdds {
				oddsLock.Lock()
				defer oddsLock.Unlock()
			}
			if o.Active != nil && *o.Active == 0 {
				odd.Model.Where(&odd.Odd{Oddid: o.Id}).Update("active", 0)
				return
			}
			//odd.Model.Where(&odd.Odd{
			//	Matchid:        match.Matchid,
			//	OddTypeId:      od.Oddtypeid,
			//	OddFieldTypeId: o.Typeid,
			//	Specialvalue:   o.Specialoddsvalue,
			//}).Update("active", 0)

			for _, of := range o.OddsField {
				//each oddfield
				//NOTE!!!! od field typeslari kapattik yeterince dbde odfieldtype var
				odf := &oddfieldType.Oddfieldtype{
					Oddtypeid: od.Oddtypeid,
					Type:      of.Type,
					Typeid:    of.Typeid,
				}
				// ////////_, err := db.DB.DB().Exec("UPDATE `odds` SET active=0 where matchid=? and oddFieldTypeId=? and oddTypeId=?",
				// ////////	*match.Matchid, *odf.Typeid, *od.Oddtypeid)
				// ////////if nil != err {
				// ////////	fmt.Println("\x1B[0m", "error updating old odds", err, "\x1B[0m")
				// ////////}
				// db.Upsert(db.DB.DB(), "oddfieldtypes", odf)
				data := &odd.Odd{
					Oddid:          o.Id,
					Matchid:        match.Matchid,
					OddFieldTypeId: odf.Typeid,
					OddTypeId:      od.Oddtypeid,
					Specialvalue:   o.Specialoddsvalue,
					Mostbalanced:   o.Mostbalanced,
					Active:         of.Active,
					OddsFieldId:    of.OddsFieldId,
				}
				if of.InnerValue != nil && *of.InnerValue != "" {
					f, err := strconv.ParseFloat(*of.InnerValue, 64)
					if nil != err {
						fmt.Println("cannot parse odd odd value=", f, err)
						return
					}
					data.Odd = &f
				}

				db.Upsert(db.DB.DB(), "odds", data)

			}
		}(*od, o)
	}
}
