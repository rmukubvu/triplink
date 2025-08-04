package handlers

import (
	"fmt"
	"time"
	"triplink/backend/database"
	"triplink/backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateNotification @Summary Create a notification
// @Description Createuser
// @Tags notifi
// @Accept json
// @Produce json
// @Param notification body models.Notificatn data"
// @Success 201 {object} models.
// @Router /notifications [post]
func CreateNotification(c *fiber.Ctx)or {
ation

	if err := c.BodyParser(&notification)nil {
		return c.Status(400).JSON(fibep{
			"se JSON",
		)
}

	// Use notification servation
	notificationService := services.GetNoce()
	createdNotification, deliveryResult, err :=ation)

	i {
{
			"error": "Could not create notificati(),
		})
	}

	// Return both notification and delivery result
	return c.Status(201).ap{
		"notification"cation,
		"delivery":     deliveryResult,
	})
}

// GetUserNotifications @Summary Get user notifns
// @Description Get all notifiuser
// @Tags notifications
on
// @Param user_id path int true "User ID"
// @Param unread
// @Success 200 {array} models.Notification
//
 {
	userID := c.Params("user_id")
	unreadOnly := c.Query("unread_only") == "true"

	query := database.DB.Whe
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}


	result := query.Order("creat


		return c.Status(500).JSON(fiber.Map{
			"error": "Could not fetch notifications",
		})
	}

	return c.JSON(notifications)
}

// MarkNotificationAsRead @Summary Maad
s read
// @Tags notifications
// @Param id path int true "Notificati
// @Success 200 {object} models.Notif
// @ad [put]
furor {
)
	var notification models.No

nil {
		return c.Status(404).JSON(
	
		})
	}

	notification.IsRead =e
	database.DB.Save(&notification)

	return c.JSON(notification)
}

as read
// @Description Mark all notifications for a user as ad
// @Tags notifications
// @Param user_id path in ID"
ce{}
// @Router /users/{user_it]
func MarkAllNotificationsAsRead(c *fib
	userID := c.Params("user_id")

	r{}).
e).
		Update("is_read", true)

	if result.Error != nil {
		rr.Map{
	ons",
})
	}

	return c.JSON(fiber.Mp{
		"message":       "All notifications marked
		"updated_count": result.RowsAffected,
	})
}

// DeleteNotification @Summary Delete
on
// @Tags notifications
// @Param id path int true "Notificati
// @Success 200 {object} map[string]sg
// @e]
fu
)
	var notification models.Notification

	if err := database.DB.First(&notification, id).E nil {
		rer.Map{
	
})
	}

	database.DB.Delete(&n
	return c.JSON(fp{
		"message": "Notification deleted succes
	})
}

// GetNotificationCounts @Summnts
ad)
// @Tags notifications
n
// @Param user_id pD"
// @Success 200 {object} map[string]interfa
// @Router /users/{user_id}/not
func GetNotification


	var totalCount, unreadCount int64

	// Get total count
n{}).
		Where("user_id = ?", userID).
		Count(&totalCount)

	//ount
	n{}).
).
		Count(&unreadCount)

	return c.JSON(map[string]interface{}{
		"total_count":  totalCount,
		"unread_count": unreadCount,
	})
}

// RegisterDeviceToken @Summartoken
// @Description Regis
//ions
// @Accept json
/oduce json

// @Param token body object true "Token data"
// @Success 200 {object} map[string]string
// @Router /users/{user_id}/notificat
func RegisterDeviceTokeor {
	userID, err := c.ParamsIntid")
	if err != nil {
		return c.Status(400).JSONer.Map{
			"error": "InvalidD",
		)
	}

dy
	var tokenData struct {
		Token      string `json:"token"`
		DeviceType string `json:"deviceType"`
		Timestamp  string `js
	}

	if err := c.BodyParser(&tokenDa{
		return c.Status(40ap{
		SON",
		})
	}

	// Validate token
	if tokenData.Token == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Token is r,
		})
	}

	// Register token w
	n()
	if err := notificationService.RegisterDeviceToil {
	

		})
	}

	return c.JSON(fiber.Map{
		"message": "Device token registered,
		"token":   tokenData.n,
	})
}

// UnregisterDeviceTce token
//ns
// @Tags notifications
/ json
son
// @Param user_id path int true "User ID"
// @Param token body object true "Token data"
// @Success 200 {object} map[string]string
// @Router /users/{usere]
func UnregisterDeviceToken(c {
	userID, err := c.ParamsInt("user_id")
	if err != nil {
		return c.Status(40r.Map{
		ID",
		})
	}

	// Parse request body
	var tokenData struct {
		Token string `json:"token"`
	}

	i{
er.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate token
	if tokenData.Token  {
		Map{
			"error": "Token is required",
	


	// Unregister token with notification service
	notificationService := services.GetNe()
	if err := notification= nil {
		return c.Status(500).JSONber.Map{
			"error": "Could not unregister device token: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
	ly",
	})
}

// GetNotificationPreferences @Summarerences
// @Description Get notification preferences for a user
// @Tags notifications
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} models.NotificationPreferences
// @Router /users/{user_id}/notification-preferences [get]
fu
_id")
	if err != nil {
		return c.Status(4{
			"error": "Invalid user ID",
		})
}

	// Get preferences from notification ser
	n

	if err != nil {
		return c.Status(500).r.Map{
			"error": "Could not get notificor(),
		})
	}

	r
}

es
// @Description Update notification preferences for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param user_id path int true
// @Param preferences body models.NotificationPrefererences"
// @Success 200 {object} modelss
// @Router /users/{uput]
fu
	userID, err := c.ParamsInt("user_id")
	
{
			"error": "Invalid user ID",
		})
	}

	// Parse request body
	var prefereces
	i {

			"error": "Cannot parse JSON",
		})
	}

	// Set user ID
	preferences.UserID = (userID)

	// Update preferencee
	no)
	if err := notificationService.Upda
		er.Map{
,
		})
	}

	return c.JSON(preferences)
}



// CreateQuoteNotification creates a noeived
func CreateQuoteNotificatio
	notification := models.Notcation{
		UserID:    shipperID,
		Title:     "New Quote Rec
		Message:   "You have rece,
		Type:      "QUOTE_RECEIVED",
		RelatedID: loadID,
	}
	
rvice()
	_, _, err := notification)
	return err
}

/

	notification := models.Notification{
		UserID:    carrierID,
		Title:     "Load Booked",
d",
		Type:      "LOAD_BOOKED",
		RelatedID: loadID,
	}
	
	not
	_ation)

}

// CreatePickupNotificaticheduled
func CreatePickupNotification(shipperID uint, loadID uint) error {
	notification := models.{
		UserID:    shipperID,
		T
	

	})
}eferences,es": prrencfe
		"preID,":     user	"user_idy",
	cessfull updated sucenceseferication prNotif":     "		"messagep{
fiber.MaN(turn c.JSO	}

	re})

		ror(), + err.Ernces: "tion preferee notificadat not up"Couldor": 
			"errfiber.Map{0).JSON(Status(50	return c.	il {
rr != nences); eID), &preferusert(ences(uintionPreferNotifica.UpdateUserionServiceotificatr := nf erce()
	ierviationSicotifces.GetN= serviionService :catfi	notiice
ation serv notificences with prefer
	// Updatet(userID)
erID = uins.UserenceID
	preft user 
	// Se	})
	}
SON",
	parse J"Cannot or": "err			fiber.Map{
SON(tatus(400).J	return c.S	nil {
r != nces); er&prefereer(Parsrr := c.Body eiferences
	onPreftificatis.Nos modeleferencey
	var prst bod reque/ Parse

	/	})
	}	 ID",
d user "Invalirror":"e	
		(fiber.Map{us(400).JSONc.Staturn et	rnil {
	rr !=  e
	if"user_id")msInt(Parar := c.D, er{
	userI error fiber.Ctx)(c *ferencesreonPNotificatirackingUpdateTnc events
fucking es for tra preferencificationdates notups referenceficationPckingNotieTra/ Updat
/
}
ences,
	})prefer: "nces	"prefere
	     userID,user_id":	".Map{
	JSON(fiberturn c.	}

	re	})
),
	rr.Error(es: " + encon preferecati get notifi notould": "C"error	
		N(fiber.Map{500).JSOatus(n c.St
		returnil {= f err !ID))
	is(uint(userncetionPrefereUserNotificarvice.GetnSetioca:= notifies, err 	preferencnService()
ioGetNotificatvices.rvice := sertionSeficaotiervice
	nation snotificm erences fro Get pref	//


		})
	}ser ID",lid unvarror": "I			"ep{
ON(fiber.Maus(400).JSStatreturn c.il {
		if err != nd")
	r_isInt("useram := c.PaerID, err
	usr {x) erroiber.Ctrences(c *fionPrefecatotifikingNunc GetTracg events
fckinnces for tra prefereficationes gets notiencferficationPreNotikingGetTrac// 
}

urn nilet	}

	r
		}
	return err
		{ nil f err !=n)
		i&notificatioDelivery(nWithtificatioateNoree.CrvictionSe notificarr :=, e, _		_ce()
cationServiNotifiservices.GetnService := otificatio
		
		n}		D: tripID,
latedIe,
			ReationTypficoti n   	Type:  sage,
		ge:   mes		Messa
	     title,		Title:erID,
	oad.ShipprID:    l	Usen{
		atio.Notific modelson :=	notificati
	 loads {= rangead :lor _, hipper
	foeach scations for eate notifi// Cr

	err
	}	return nil {
	or; err != &loads).Errind( tripID).F= ?",id e("trip_se.DB.Wherdataba err := Load
	ifdels.movar loads []	 trip
on thisoads // Get all lerror {
	 string) ficationTypeotig, nage string, messitle strinD uint, tnTrip(tripIrsOlShippeNotifyAl
func  triploads on apers with ip all shtions tocatifirip sends noippersOnTotifyAllSh
}

// N	return errion)
notificatvery(&WithDelitionotificaeNvice.CreationSer= notificat err :_, _,e()
	rvicSeicationtNotifices.Gevice := servificationSer
	not
	}
	 tripID,tedID:",
		RelaON_UPDATE "LOCATI   e:  n,
		Typ locatioed " +achs rehipment ha:   "Your sge	Messa
	",Updatetion     "Loca: 
		Title shipperID,UserID:   {
		ification models.Notfication :=notior {
	tring) errtion s locaD uint,int, tripIipperID uion(shificatateNottionUpdcaeLo
func Creation updatest locatignificanfor sfication ates a notiation creictifionUpdateNocat
// CreateLo
}
	return errcation)
&notifiDelivery(ionWithNotificatce.CreateicationServiif:= notrr 	_, _, ee()
ionServictificatetNo services.GService :=otification
	
	n
	}ID,dID: load		RelateHANGED",
S_CSTATU   "LOAD_Type:   
		ssage, me	Message:  te",
	us Upda Statoad"LTitle:     	,
	IDhipper  srID:  on{
		Uses.Notificati= modelation :tificno

	)"
	}kingRef + "f: " + boo(Re " essage +=" {
		m != " bookingRef}

	ifs
	atuewSt+ nted to " da upen betatus hasload sge = "Your 
		messae == "" { messag]
	ifwStatusMessages[ne:= statusge 
	messa",
	}
 attentiont requiresd thaour loae with ysu is an is     "There  XCEPTION": ,
		"Edelivered"ccessfully s been suYour load ha"        D":	"DELIVEREvery",
	t for deliis our load ": "YouOR_DELIVERY	"OUT_Fsit",
	ranin tcurrently is our load  "Y      ":TRANSIT"IN_t",
		sin trans now i i and up picked been hasr load"You:        P"PICKED_U
		"heduled",s been schap d pickuloaD": "Your CHEDULE_SICKUP{
		"Ptring]stringes := map[statusMessagerror {
	sg)  strin bookingReftring, newStatus st,D uin, loadIntrID uiion(shippeficatsChangeNotidStatureateLoanc Changes
fuatus chen load stion w a notificatn createsNotificatioStatusChangeLoad

// Creatern err
}ion)
	retunotificaty(&WithDeliveronatireateNotificervice.CcationS notifi :=_, err, 
	_ice()ervcationSs.GetNotifie := serviceationServic
	notific	,
	}
dID: tripIDlate",
		ReA_UPDATEDe:      "ETTyp,
		4 PM")006 at 3:0("Jan 2, 2matFor" + newETA.pdated to n uime has bee tivalated arrs estimipment'sh   "Your essage:ed",
		MUpdat     "ETA 
		Title:pperID,  shi:  
		UserIDtion{otifica models.Nion :=	notificat) error {
ime.TimeTA tint, newEripID unt, tui(shipperID ontiNotificaTAUpdatenc CreateEpdated
fuis uA  ETwhentification es a noon creatteNotificatiUpdaeETACreat
}

// rrreturn eon)
	otificatiy(&nelivercationWithDtifiice.CreateNoionServotificat _, err := n
	_,e()tionServictNotificas.Ge= serviceService :tification
	
	noripID,
	}: tRelatedID
		DELAYED",    "TRIP_pe:  
		Tyssage,  me	Message: ",
	 Delayed"Shipment     		Title:D,
serI    uD:
		UserItion{ificamodels.Noton := tificati

	noon
	}+ rease to "  " du	message +="" {
	= f reason !inutes)
	iayM", del %d minuteselayed byis dt ipmen("Your sh fmt.Sprintfsage :=	mes) error {
eason stringes int, r delayMinutripID uint,rID uint, tfication(useNotiateDelayunc Cres delayed
fip ihen a trion wificat a not createstificationayNoel
// CreateDn err
}
	returtion)
icavery(&notifWithDeliotificationce.CreateNationServificoti err := n_,
	_, ce()viSercationtNotifirvices.Ge := seationServicefic	noti	}
	
,
ID: tripIDelated
		RP_ARRIVED",RI   "Te:   yp	T	,
prepared."being very is n + ". Delitioloca" + ed at as arrivhipment h"Your s	Message:   	
 Arrived", "Trip	Title:    erID,
	shipp:    
		UserIDation{Notificn := models.notificatior {
	erron string) tioint, locaipID ut, trinerID u(shippionalNotificatipArrivateTrunc Cretion
fat destinas trip arriveen a whation ficeates a notication crivalNotifieTripArr

// Createturn err
}	rcation)
otifielivery(&ncationWithDfiNotireateionService.Ccatfir := noti_, er, e()
	_rviconSeificatitNotces.Ge := serviiceervcationSnotifi
	
	ID,
	}D: tripRelatedI",
		ED"TRIP_DEPART:      peTy",
		 in transit.nowed and is art dep " hasrName + + carrieh "witpment ur shie:   "Yo	Messag
	ed","Trip Departtle:     	TirID,
	:    shippeID
		Useron{tis.Notifica := modelicationifot	n error {
e string)rrierNamcaint, ID unt, triprID uiion(shippeaticNotifpDeparturereateTrifunc Cdeparts
en a trip whon notificatites a creaon NotificatipDepartureteTri

// Crea functionsicationcific notif-specking/ Traerr
}

/	return 
n)tificatioy(&nonWithDeliverificatioeNotervice.CreatcationS notifi :=	_, _, errervice()
onSatiotifices.GetN service :=Servicon	notificati}
	
ID,
	ad lo	RelatedID:	",
D_DELIVERED:      "LOAype.",
		Tierhe carrreview te red. Pleasdeliveully n successfeead has b"Your lossage:   		Meered",
"Load Deliv    e: 	Titl
	shipperID,	UserID:    on{
	catifi models.Notin :=notificatiorror {
	int) e uadIDID uint, lohippertion(sificaryNotveCreateDeli
func livered is de loadwhencation s a notifi createcationfieliveryNoti// CreateDerr
}

turn reation)
	ificlivery(&notnWithDeotificatiovice.CreateNficationSernoti, err := 	_, _e()
ervicificationSvices.GetNotsere := ictionServtifica
	no
	}
	D: loadID,dIelate,
		RULED"CKUP_SCHED"PI    ype:  		T