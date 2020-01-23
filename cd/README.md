# Continuous deployment

CD consists of the following steps:

* Some updates are pushed to master branch
* Github Actions [rebuild images](../.github/workflows/main.yml)
* New prerelease is created on github **manually**
* Github Actions [sends webhooks](../.github/workflows/release.yml) to [WSGI Application](#wsgi-application)
* WSGI App [fetches new docker image and recreate the docker container](deploy.example.sh)
* Deployment is checked **manually**
* Finally, release is published **manually**

# WSGI App

* Configure access to github registry with [token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line):

```sh
docker login docker.pkg.github.com -u GITHUB_USERNAME -p GITHUB_TOKEN 
```

* [install uwsgi](https://uwsgi-docs.readthedocs.io/en/latest/WSGIquickstart.html#installing-uwsgi-with-python-support). 

* Make `deploy.sh` file out of [deploy.example.sh](deploy.example.sh).

* Deploy WSGI App:

```sh
uwsgi --http :9090 --wsgi-file deploy-uwsgi.py  &> hound.logs &
```

* [Create secret](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/creating-and-using-encrypted-secrets#creating-encrypted-secrets) `DEPLOYMENT_WEBHOOK` with the url to your WSGI App
