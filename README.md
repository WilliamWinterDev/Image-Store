# Image Store
A private image store service which can will be used for hosting sharable screenshots integrated with ShareX. ShareX will upload images to the Go API and a random file name will be generated and stored in a K/V store. Whens someone queries the image it will look into the K/V storage to find the image.

An admin panel will be created to manage the K/V store so that images can be removed on request.
