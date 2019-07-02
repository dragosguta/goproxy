package main

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

// CognitoAppClient is an interface for working with AWS Cognito
type CognitoAppClient struct {
	Region           string
	UserPoolID       string
	ClientID         string
	WellKnownJWKs    *jwk.Set
	IdentityProvider *cognitoidentityprovider.CognitoIdentityProvider
}

// CognitoAppClientConfig handles the pool, region, and client
type CognitoAppClientConfig struct {
	Region   string `json:"region"`
	PoolID   string `json:"poolId"`
	ClientID string `json:"clientId"`
}

// CognitoToken defines a token struct for JSON responses from Cognito TOKEN endpoint
type CognitoToken struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
}

// CognitoAccessTokenClaims on the cognito access token
type CognitoAccessTokenClaims struct {
	AuthTime int64  `json:"auth_time"`
	ClientID string `json:"client_id"`
	EventID  string `json:"event_id"`
	Exp      int64  `json:"exp"`
	Iss      string `json:"iss"`
	Jti      string `json:"jti"`
	Scope    string `json:"scope"`
	Sub      string `json:"sub"`
	TokenUse string `json:"token_use"`
	Username string `json:"username"`
}

// NewCognitoAppClient returns a new CognitoAppClient interface configured for the given Cognito user pool and client
func NewCognitoAppClient(cfg *CognitoAppClientConfig) (*CognitoAppClient, error) {
	var err error

	// Set up identity provider
	svc := cognitoidentityprovider.New(session.New(), &aws.Config{Region: aws.String(cfg.Region)})

	c := &CognitoAppClient{
		Region:           cfg.Region,
		UserPoolID:       cfg.PoolID,
		ClientID:         cfg.ClientID,
		IdentityProvider: svc,
	}

	// Set the well known JSON web token key sets
	err = c.getWellKnownJWTKs()
	if err != nil {
		log.Println("error getting well known JWTKs", err)
	}

	return c, err
}

// getWellKnownJWTKs gets the well known JSON web token key set for this client's user pool
func (c *CognitoAppClient) getWellKnownJWTKs() error {
	// https://cognito-idp.<region>.amazonaws.com/<pool_id>/.well-known/jwks.json
	var buffer bytes.Buffer
	buffer.WriteString("https://cognito-idp.")
	buffer.WriteString(c.Region)
	buffer.WriteString(".amazonaws.com/")
	buffer.WriteString(c.UserPoolID)
	buffer.WriteString("/.well-known/jwks.json")
	wkjwksURL := buffer.String()
	buffer.Reset()

	// Use this cool package
	set, err := jwk.Fetch(wkjwksURL)
	if err == nil {
		c.WellKnownJWKs = set
	} else {
		log.Println("there was a problem getting the well known JSON web token key set")
		log.Println(err)
	}
	return err
}

func (c *CognitoAppClient) authenticate(token string) (User, error) {
	parsedToken, err := parseJWT(token, c.WellKnownJWKs)
	if err != nil {
		return User{authenticated: false}, err
	}

	validatedToken, err := validateJWT(parsedToken, c.ClientID)
	if err != nil {
		return User{authenticated: false}, err
	}

	username := validatedToken.Claims.(jwt.MapClaims)["username"].(string)
	if username == "" {
		return User{authenticated: false}, err
	}

	parameters := &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: &c.UserPoolID,
		Username:   &username,
	}

	req, resp := c.IdentityProvider.AdminGetUserRequest(parameters)

	err = req.Send()
	if err != nil {
		return User{authenticated: false}, err
	}

	attributes := UserAttributes{
		Enabled:          *resp.Enabled,
		CreatedDate:      *resp.UserCreateDate,
		LastModifiedDate: *resp.UserLastModifiedDate,
		Username:         *resp.Username,
		Status:           *resp.UserStatus,
	}

	for _, value := range resp.UserAttributes {
		attributes.Attributes = append(attributes.Attributes, UserAttributeField{
			Name:  ToLowerCamel(*value.Name),
			Value: *value.Value,
		})
	}

	return User{
		authenticated: true,
		claims:        validatedToken.Claims,
		attributes:    attributes,
	}, nil
}
