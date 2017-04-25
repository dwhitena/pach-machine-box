# Building a "Trump Finder" with Pachyderm and Machine Box

This tutorial will walk you through building a "Trump Finder," a data pipeline that will recognize Donald Trump's face in images and tag them.  In face, the data pipeline can do even more than that.  Generally, you can train this pipeline to identify any of number people and tag those people in images.  

## Getting up and running with Pachyderm

You can experiment with this pipeline locally using a quick [local installation of Pachyderm](http://docs.pachyderm.io/en/latest/getting_started/local_installation.html).  Alternatively, you can quickly spin up a real Pachyderm cluster in any one of the popular cloud providers.  Check out the [Pachyderm docs](http://docs.pachyderm.io/en/latest/deployment/deploy_intro.html) for more details on deployment.

Once deployed, you will be able to use the Pachyderm’s `pachctl` CLI tool to create data repositories and build the facial recognition pipeline.

## Getting up and running with Machine Box

We will be utilizing (Machine Box)[https://machinebox.io/] to learn and identify faces within images (e.g., Donald Trump).  In particular, we will be utilizing their "facebox," which is a pre-built machine learning "battery" that you can plug into any system to enable facial recognition.  When facebox is launched you will immediately have a JSON API will will allow you to teach a model certain faces, export that model, import the model, and identify faces in images with the model.

To run this tutorial you will need to have a Machine Box API key, which can be obtained [here](https://machinebox.io/).  Once you have the key, make sure and export it to an environmental variable:

```
➔ export MB_KEY=<your MB API key>
```

## Creating input data repositories

The input and output of any pipeline stage in a Pachyderm pipeline is a "data repository."  In these repos, Pachyderm versions your data (think "git for data") such that you always know the state of your data at any time and for any run.  Thus, you can maintain reproducibilty.

As a first step towards our "Trump Finder," we need to create our input data repositories for our:

- training data - which will include images of faces for Machine Box to learn
- unidentified data - which will include images with unidentified faces in them
- label data - which will include labels for our pipeline to use when tagging identified faces in images

To create these repo, run:

```
➔ pachctl create-repo training
➔ pachctl create-repo unidentified
➔ pachctl create-repo labels
➔ pachctl list-repo
NAME                CREATED             SIZE                
labels              3 seconds ago       0 B                 
unidentified        11 seconds ago      0 B                 
training            17 seconds ago      0 B                 
➔
```

## Getting our input data into Pachyderm

Next we need to actually put our training images into the `training` data repository.  We will add some trump faces under [data/train/faces1](data/train/faces1) to our repo to start with:

```
➔ cd data/train/faces1/
➔ ls
trump1.jpg  trump2.jpg  trump3.jpg  trump4.jpg  trump5.jpg
➔ pachctl put-file training master -c -r -f .
➔ pachctl list-repo
NAME                CREATED             SIZE                
training            5 minutes ago       486.2 KiB           
labels              5 minutes ago       0 B                 
unidentified        5 minutes ago       0 B                 
➔ pachctl list-file training master
NAME                TYPE                SIZE                
trump1.jpg          file                78.98 KiB           
trump2.jpg          file                334.5 KiB           
trump3.jpg          file                11.63 KiB           
trump4.jpg          file                27.45 KiB           
trump5.jpg          file                33.6 KiB            
➔
```

We will also put a label image for Trump's face and a couple of unidentified images into their respective inputs:

```
➔ cd ../../labels/
➔ ls
clinton.jpg  trump.jpg
➔ pachctl put-file labels master -c -r -f .
➔ cd ../unidentified/
➔ ls
image1.jpg  image2.jpg
➔ pachctl put-file unidentified master -c -r -f .
➔ pachctl list-repo
NAME                CREATED             SIZE                
unidentified        7 minutes ago       540.4 KiB           
labels              7 minutes ago       15.44 KiB           
training            7 minutes ago       486.2 KiB           
➔
```

## Creating a pipeline to train a facial recognition model

As mentioned we will use machine box to train our a facial recognition model to learn Trump's face (along with others).  Machine Box provides a Docker image `machinebox/facebox` for this purpose.  We have create another [Docker image](pachfacebox/Dockerfile) based on `machinebox/facebox`, which adds `cURL` to the image such that we can call the facebox API programmatically in our pipeline.

We also have created a [Pachyderm pipeline specification](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html) for the training stage of our pipeline.  This can be found [here](pipelines/train.json).  It tells Pachyderm to use our `pachfacebox` image to process images in our `training` repo with the facebox endpoint for teaching the model a face.

This pipeline is then created as follows:

```
➔ cd ../../pipelines/
➔ ls
create-MB-pipeline.sh  identify.json  tag.json  train.json
➔ ./create-MB-pipeline.sh train.json 
➔ pachctl list-pipeline
NAME                INPUT               OUTPUT              STATE               
model               training            model/master        running    
➔ pachctl list-job
ID                                   OUTPUT COMMIT STARTED       DURATION RESTART PROGRESS STATE            
3425a7a0-543e-4e2a-a244-a3982c527248 model/-       9 seconds ago -        1       0 / 1    running 
➔
```

You will notice that immediately Pachyderm has started a job to process the face images that we previously put into the `training` repo.  Eventually, this job will succeed (note the first time you run this Pachyderm will have to pull the images, which might take a few minutes depends on your connection to Docker Hub):

```
➔ pachctl list-job
ID                                   OUTPUT COMMIT                          STARTED       DURATION  RESTART PROGRESS STATE            
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a 5 minutes ago 5 minutes 1       1 / 1    success 
➔ pachctl list-repo
NAME                CREATED             SIZE                
model               5 minutes ago       4.118 KiB           
unidentified        18 minutes ago      540.4 KiB           
labels              18 minutes ago      15.44 KiB           
training            19 minutes ago      486.2 KiB           
➔ pachctl list-file model master
NAME                TYPE                SIZE                
state.facebox       file                4.118 KiB           p
➔
```

The output of the job is a state file which represents our trained model.  We will utilize this is a next pipeline stage to identify faces.

## Creating a pipeline to identify faces in unidentified images

We have also created a machine box based pipeline specification, [identify.json](pipelines/identify.json), which utilizes Machine Box to identify faces within images in the `unidentified` data repository.  Similar to above, we can create this pipeline as follows:

```
➔ ./create-MB-pipeline.sh identify.json 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                          STARTED       DURATION  RESTART PROGRESS STATE            
281d4393-05c8-44bf-b5de-231cea0fc022 identify/-                             6 seconds ago -         0       0 / 2    running 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a 8 minutes ago 5 minutes 1       1 / 1    success 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED            DURATION   RESTART PROGRESS STATE            
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 About a minute ago 53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    9 minutes ago      5 minutes  1       1 / 1    success 
➔ pachctl list-repo
NAME                CREATED              SIZE                
identify            About a minute ago   1.932 KiB           
model               10 minutes ago       4.118 KiB           
unidentified        23 minutes ago       540.4 KiB           
labels              23 minutes ago       15.44 KiB           
training            24 minutes ago       486.2 KiB           
➔ pachctl list-file identify master
NAME                TYPE                SIZE                
image1.json         file                1.593 KiB           
image2.json         file                347 B               
➔ pachctl get-file identify master image1.json
{
	"success": true,
	"facesCount": 13,
	"faces": [
		{
			"rect": {
				"top": 199,
				"left": 677,
				"width": 107,
				"height": 108
			},
			"matched": false
		},
		{
			"rect": {
				"top": 96,
				"left": 1808,
				"width": 89,
				"height": 90
			},
			"matched": false
		},
		{
			"rect": {
				"top": 163,
				"left": 509,
				"width": 108,
				"height": 108
			},
			"matched": false
		},
		{
			"rect": {
				"top": 186,
				"left": 2186,
				"width": 89,
				"height": 89
			},
			"matched": false
		},
		{
			"rect": {
				"top": 166,
				"left": 1638,
				"width": 90,
				"height": 89
			},
			"matched": false
		},
		{
			"rect": {
				"top": 116,
				"left": 1453,
				"width": 107,
				"height": 107
			},
			"matched": false
		},
		{
			"rect": {
				"top": 405,
				"left": 1131,
				"width": 89,
				"height": 89
			},
			"matched": false
		},
		{
			"rect": {
				"top": 206,
				"left": 1300,
				"width": 90,
				"height": 89
			},
			"matched": false
		},
		{
			"rect": {
				"top": 176,
				"left": 1957,
				"width": 90,
				"height": 89
			},
			"matched": false
		},
		{
			"rect": {
				"top": 495,
				"left": 1462,
				"width": 62,
				"height": 62
			},
			"matched": false
		},
		{
			"rect": {
				"top": 175,
				"left": 975,
				"width": 108,
				"height": 108
			},
			"id": "58ff31510f7707a01fb3e2f4d39f26dc",
			"name": "trump",
			"matched": true
		},
		{
			"rect": {
				"top": 1158,
				"left": 2181,
				"width": 62,
				"height": 63
			},
			"matched": false
		},
		{
			"rect": {
				"top": 544,
				"left": 1647,
				"width": 75,
				"height": 75
			},
			"matched": false
		}
	]
}
```

Awesome! We have identified Donald Trump in this image.  

## Tagging images at the identified locations

To make the above identification a little more visual, we have created a [Go program](tagimage/main.go) that overlays a label image from the `labels` repository at the locations of identified faces.  There is another pipeline specification that goes along with this named [tag.json](pipelines/tag.json).  To create the tagging pipeline:

```
➔ pachctl create-pipeline -f tag.json 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED        DURATION   RESTART PROGRESS STATE            
cd284a28-6c97-4236-9f6d-717346c60f24 tag/-                                     2 seconds ago  -          0       0 / 2    running 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 5 minutes ago  53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    13 minutes ago 5 minutes  1       1 / 1    success 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED        DURATION   RESTART PROGRESS STATE            
cd284a28-6c97-4236-9f6d-717346c60f24 tag/ae747e8032704b6cae6ae7bba064c3c3      25 seconds ago 11 seconds 0       2 / 2    success 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 5 minutes ago  53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    14 minutes ago 5 minutes  1       1 / 1    success 
➔ pachctl list-repo
NAME                CREATED             SIZE                
tag                 30 seconds ago      591.3 KiB           
identify            5 minutes ago       1.932 KiB           
model               14 minutes ago      4.118 KiB           
unidentified        27 minutes ago      540.4 KiB           
labels              27 minutes ago      15.44 KiB           
training            27 minutes ago      486.2 KiB           
➔ pachctl list-file tag master
NAME                TYPE                SIZE                
tagged_image1.jpg   file                557 KiB             
tagged_image2.jpg   file                34.35 KiB           
➔
```

The tagged images look like this:

![alt text](tagged_images1.jpg)

## Teaching the model to learn, identify, and tag other faces

Now that we have this pipeline built, it is super easy to teach the model more faces.  Moreover, when we add new face images to the `training` repo, Pachyderm will automatically update our model, identification, and tagging.  We don't have to worry about running everything manually.  The result are kept in sync with changes in the input data.  Let's teach our model Hillary Clinton's face and update our tagging:

```
➔ cd ../data/train/faces2/
➔ ls
clinton1.jpg  clinton2.jpg  clinton3.jpg  clinton4.jpg
➔ pachctl put-file training master -c -r -f .
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED        DURATION   RESTART PROGRESS STATE            
56e24ac0-0430-4fa4-aa8b-08de5c1884db model/-                                   4 seconds ago  -          0       0 / 1    running 
cd284a28-6c97-4236-9f6d-717346c60f24 tag/ae747e8032704b6cae6ae7bba064c3c3      6 minutes ago  11 seconds 0       2 / 2    success 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 11 minutes ago 53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    20 minutes ago 5 minutes  1       1 / 1    success 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED            DURATION   RESTART PROGRESS STATE            
8a7961b7-1085-404a-b0ee-66034fae7212 identify/-                                48 seconds ago     -          0       1 / 2    running 
56e24ac0-0430-4fa4-aa8b-08de5c1884db model/002f16b63a4345a4bc6bdf5510c9faac    About a minute ago 19 seconds 0       1 / 1    success 
cd284a28-6c97-4236-9f6d-717346c60f24 tag/ae747e8032704b6cae6ae7bba064c3c3      7 minutes ago      11 seconds 0       2 / 2    success 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 13 minutes ago     53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    21 minutes ago     5 minutes  1       1 / 1    success 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED                DURATION   RESTART PROGRESS STATE            
6aa6c995-58ce-445d-999a-eb0e0690b041 tag/-                                     Less than a second ago -          0       1 / 2    running 
8a7961b7-1085-404a-b0ee-66034fae7212 identify/1bc94ec558e44e0cb45ed5ab7d9f9674 54 seconds ago         54 seconds 0       2 / 2    success 
56e24ac0-0430-4fa4-aa8b-08de5c1884db model/002f16b63a4345a4bc6bdf5510c9faac    About a minute ago     19 seconds 0       1 / 1    success 
cd284a28-6c97-4236-9f6d-717346c60f24 tag/ae747e8032704b6cae6ae7bba064c3c3      8 minutes ago          11 seconds 0       2 / 2    success 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 13 minutes ago         53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    21 minutes ago         5 minutes  1       1 / 1    success 
➔ pachctl list-job
ID                                   OUTPUT COMMIT                             STARTED            DURATION   RESTART PROGRESS STATE            
6aa6c995-58ce-445d-999a-eb0e0690b041 tag/7cbd2584d4f0472abbca0d9e015b9829      5 seconds ago      1 seconds  0       2 / 2    success 
8a7961b7-1085-404a-b0ee-66034fae7212 identify/1bc94ec558e44e0cb45ed5ab7d9f9674 59 seconds ago     54 seconds 0       2 / 2    success 
56e24ac0-0430-4fa4-aa8b-08de5c1884db model/002f16b63a4345a4bc6bdf5510c9faac    About a minute ago 19 seconds 0       1 / 1    success 
cd284a28-6c97-4236-9f6d-717346c60f24 tag/ae747e8032704b6cae6ae7bba064c3c3      8 minutes ago      11 seconds 0       2 / 2    success 
281d4393-05c8-44bf-b5de-231cea0fc022 identify/287fc78a4cdf42d89142d46fb5f689d9 13 minutes ago     53 seconds 0       2 / 2    success 
3425a7a0-543e-4e2a-a244-a3982c527248 model/1b9c158e33394056a18041a4a86cb54a    21 minutes ago     5 minutes  1       1 / 1    success 
➔ pachctl list-file tag master
NAME                TYPE                SIZE                
tagged_image1.jpg   file                557 KiB             
tagged_image2.jpg   file                36.03 KiB           
➔
```

Now our tagged images look like this:

![alt text](tagged_images2.jpg)


