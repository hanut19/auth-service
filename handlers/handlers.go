package handlers

import (
	"auth-service/common"
	"auth-service/constants"
	"auth-service/database"
	"auth-service/helpers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"auth-service/logger"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"gopkg.in/mgo.v2/bson"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	var reqUserparams common.User
	err := json.NewDecoder(r.Body).Decode(&reqUserparams)
	if err != nil {
		logger.ErrorLogger.Printf("SignUp: Error in reading payload.")
		res := common.APIResponse{
			StatusCode: 400,
			Message:    "Error in reading payload.",
			IsError:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	v := validator.New()
	err = v.Struct(reqUserparams)
	if err != nil {
		logger.ErrorLogger.Printf("SignUp: Error in validating Input %s", err.Error())
		res := common.APIResponse{
			StatusCode: 400,
			Message:    err.Error(),
			IsError:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}
	allRoles := constants.GetRole()
	reqRole := string(reqUserparams.Role)
	if !helpers.ArrayValCheck(allRoles, reqRole) {
		logger.ErrorLogger.Printf("SignUp: Value of role is not correct")
		res := common.APIResponse{
			StatusCode: 400,
			Message:    "Value of Role is not correct",
			IsError:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return

	}

	//fmt.Println(reqUserparams.Email)
	var respUser common.User
	dbConn := database.Connection()
	usersession := dbConn.Database("usercatalog").Collection("userprofile")
	_ = usersession.FindOne(context.TODO(), bson.M{"email": reqUserparams.Email}).Decode(&respUser)
	defer database.CloseClientDB(dbConn)
	if respUser.Email == "" {
		logger.ErrorLogger.Printf("SignUp: Email ID cannot be empty")
		hashPassword, err := helpers.GeneratehashPassword(reqUserparams.Password)
		if err == nil {
			reqUserparams.Password = hashPassword
		}
		result, err := usersession.InsertOne(context.TODO(), &reqUserparams)
		if err != nil {
			mesg := fmt.Sprintf("Inseration failed with error %s", err.Error())
			//logger.Error(mesg)
			logger.ErrorLogger.Printf(mesg)
			fmt.Println(mesg)
			res := common.APIResponse{
				StatusCode: 500,
				Message:    mesg,
				IsError:    true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}
		res := common.APIResponse{
			StatusCode: 201,
			Message:    "User Created Sucessfully!!",
			Result:     result,
		}
		logger.InfoLogger.Printf("User Created Sucessfully")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
		return

	} else {
		mesg := "Duplicate Record"
		fmt.Println(mesg)
		logger.ErrorLogger.Printf("SignUp: Duplicate Record")
		//logger.Info(mesg)
		res := common.APIResponse{
			IsError:    true,
			StatusCode: 409,
			Message:    "Email Allready Exists. Duplicate Record !!",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(res)
		return
	}
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	dbConn := database.Connection()
	usersession := dbConn.Database("usercatalog").Collection("userprofile")
	defer database.CloseClientDB(dbConn)
	var authDetails common.Authentication

	err := json.NewDecoder(r.Body).Decode(&authDetails)
	if err != nil {
		logger.ErrorLogger.Printf("SignIn: Error in reading payload")
		res := common.APIResponse{
			StatusCode: 400,
			Message:    "Error in reading payload.",
			IsError:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(res)
		return
	}

	var authUser common.User
	_ = usersession.FindOne(context.TODO(), bson.M{"email": authDetails.Email}).Decode(&authUser)

	if authUser.Email == "" {
		logger.ErrorLogger.Printf("SignIn: Username or Password not matched. ")
		res := common.APIResponse{
			StatusCode: 200,
			Message:    "Username or Password is incorrect",
			IsError:    true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	check := helpers.CheckPasswordHash(authDetails.Password, authUser.Password)

	if !check {
		logger.ErrorLogger.Printf("SignIn: Username or Password is incorrect")
		res := common.APIResponse{
			StatusCode: 200,
			Message:    "Username or Password is incorrect",
			IsError:    true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	validToken, err := helpers.GenerateJWT(authUser.Email, authUser.Role)
	if err != nil {
		logger.ErrorLogger.Printf("SignIn: Error in Generating Token %s", err.Error())
		res := common.APIResponse{
			StatusCode: 200,
			Message:    "Failed to generate token",
			IsError:    true,
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		return
	}

	var token common.Token
	token.Email = authUser.Email
	token.Role = authUser.Role
	token.TokenString = validToken
	//DB
	var respUsertoken common.Token
	//dbConn = database.Connection()
	usersession = dbConn.Database("usercatalog").Collection("usertoken")
	_ = usersession.FindOne(context.TODO(), bson.M{"email": authUser.Email}).Decode(&respUsertoken)
	fmt.Println(respUsertoken)
	if respUsertoken.Email == "" {
		_, err := usersession.InsertOne(context.TODO(), &token)
		if err != nil {
			//mesg := fmt.Sprintf("Inseration failed with error %s", err.Error())
			mesg := fmt.Sprintf("Failed to generate token %s", err.Error())
			fmt.Println(mesg)
			logger.ErrorLogger.Printf("SignIn:" + mesg)
			res := common.APIResponse{
				StatusCode: 500,
				Message:    mesg,
				IsError:    true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}
	} else {
		change := bson.M{"$set": bson.M{"tokenstring": validToken}}
		_, err := usersession.UpdateOne(context.TODO(), bson.M{"email": authUser.Email}, change)
		if err != nil {
			mesg := fmt.Sprintf("Failed to generate token %s", err.Error())
			fmt.Println(mesg)
			logger.ErrorLogger.Printf("SignIn :" + mesg)
			res := common.APIResponse{
				StatusCode: 500,
				Message:    mesg,
				IsError:    true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}
	}

	//End DB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HOME PUBLIC INDEX PAGE"))
}

func AdminIndex(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Role") != constants.ADMIN {
		w.Write([]byte("Not authorized."))
		return
	}
	w.Write([]byte("Welcome, Admin."))
}

func UserIndex(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Role") != constants.USER {
		w.Write([]byte("Not Authorized."))
		return
	}
	w.Write([]byte("Welcome, User"))
}

//check whether user is authorized or not
func IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			logger.ErrorLogger.Printf("Auth: No Token Found")
			var err common.Error
			err = helpers.SetError(err, "No Token Found")
			json.NewEncoder(w).Encode(err)
			return
		}
		reqToken = strings.TrimSpace(splitToken[1])
		dbConn := database.Connection()
		var respUsertoken common.Token
		usersession := dbConn.Database("usercatalog").Collection("usertoken")
		_ = usersession.FindOne(context.TODO(), bson.M{"tokenstring": reqToken}).Decode(&respUsertoken)
		if respUsertoken.TokenString == "" {
			logger.ErrorLogger.Printf("Auth: Your Token is not vaild")
			var err common.Error
			err = helpers.SetError(err, "Your Token is not vaild.")
			json.NewEncoder(w).Encode(err)
			return
		}

		var mySigningKey = []byte(constants.SECRETKEY)

		token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing token.")
			}
			return mySigningKey, nil
		})
		if err != nil {
			logger.ErrorLogger.Printf("Auth: Your Token has been expired")
			var err common.Error
			err = helpers.SetError(err, "Your Token has been expired.")
			json.NewEncoder(w).Encode(err)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == constants.ADMIN {
				r.Header.Set("Role", constants.ADMIN)
				handler.ServeHTTP(w, r)
				return
			} else if claims["role"] == constants.USER {
				r.Header.Set("Role", constants.USER)
				handler.ServeHTTP(w, r)
				return
			}
		}
		var reserr common.Error
		logger.ErrorLogger.Printf("Auth: Your are not authorized")
		reserr = helpers.SetError(reserr, "Not Authorized.")
		json.NewEncoder(w).Encode(err)
	}
}
