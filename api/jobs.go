package api

//const helloMessage = "Hello, World!"
//
//type HelloResponse struct {
//	Message string `json:"message,omitempty"`
//}

// CreateJobHandler is a handler that inserts a job into mongoDB with a newly generated ID and links
//func (api *API) CreateJobHandler(w http.ResponseWriter, req *http.Request) {
//	ctx := req.Context()
//	hColID := ctx.Value(handlers.CollectionID.Context())
//	logdata := log.Data{
//		handlers.CollectionID.Header(): hColID,
//		"request-id":                   ctx.Value(netreq.RequestIdKey),
//	}
//
//	newImageRequest := &models.Image{}
//	if err := ReadJSONBody(ctx, req.Body, newImageRequest); err != nil {
//		handleError(ctx, w, err, logdata)
//		return
//	}
//
//	id := NewID()
//
//	// generate new image from request, mapping only allowed fields at creation time (model newImage in swagger spec)
//	// the image is always created in 'created' state, and it is assigned a newly generated ID
//	newImage := models.Image{
//		ID:           id,
//		CollectionID: newImageRequest.CollectionID,
//		State:        newImageRequest.State,
//		Filename:     newImageRequest.Filename,
//		License:      newImageRequest.License,
//		Links:        api.createLinksForImage(id),
//		Type:         newImageRequest.Type,
//	}

//	// generic image validation
//	if err := newImage.Validate(); err != nil {
//		handleError(ctx, w, err, logdata)
//		return
//	}
//
//	// Check provided image state supplied is correct
//	if newImage.State != models.StateCreated.String() {
//		handleError(ctx, w, apierrors.ErrImageBadInitialState, logdata)
//		return
//	}
//	mage.CollectionID == "" {
//		handleError(ctx, w, apierrors.ErrImageNoCollectionID, logdata)
//		return
//	}
//
//	log.Event(ctx, "storing new image", log.INFO, log.Data{"image": newImage})
//
//	// Upsert image in MongoDB
//	if err := api.mongoDB.UpsertImage(req.Context(), newImage.ID, &newImage); err != nil {
//		handleError(ctx, w, err, logdata)
//		return
//	}
//
//	w.WriteHeader(http.StatusCreated)
//	if err := WriteJSONBody(ctx, newImage, w, logdata); err != nil {
//		handleError(ctx, w, err, logdata)
//		return
//	}
//	log.Event(ctx, "successfully created image", log.INFO, logdata)
//}

// HelloHandler returns function containing a simple hello world example of an api handler
//func HelloHandler(ctx context.Context) http.HandlerFunc {
//	log.Event(ctx, "api contains example endpoint, remove hello.go as soon as possible", log.INFO)
//	return func(w http.ResponseWriter, req *http.Request) {
//		ctx := req.Context()
//
//		response := HelloResponse{
//			Message: helloMessage,
//		}
//
//		w.Header().Set("Content-Type", "application/json; charset=utf-8")
//		jsonResponse, err := json.Marshal(response)
//		if err != nil {
//			log.Event(ctx, "marshalling response failed", log.Error(err), log.ERROR)
//			http.Error(w, "Failed to marshall json response", http.StatusInternalServerError)
//			return
//		}
//
//		_, err = w.Write(jsonResponse)
//		if err != nil {
//			log.Event(ctx, "writing response failed", log.Error(err), log.ERROR)
//			http.Error(w, "Failed to write http response", http.StatusInternalServerError)
//			return
//		}
//	}
//}
