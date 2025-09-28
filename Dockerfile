FROM alpine:latest 
RUN apk --no-cache add ca-certificates
WORKDIR /app  
COPY . . 
EXPOSE 8080
CMD "./main"

